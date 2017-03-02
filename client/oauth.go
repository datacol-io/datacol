package client

import (
  "fmt"
  "os"
  "log"
  "net"
  "net/http"
  "context"
  "encoding/base64"
  "encoding/json"

  "github.com/pkg/errors"
  "github.com/skratchdot/open-golang/open"
  goauth2 "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  crmgr "google.golang.org/api/cloudresourcemanager/v1"
  iam "google.golang.org/api/iam/v1"

  "github.com/dinesh/rz/client/helper"
)

const (
  googleOauth2ClientID     = "992213213700-ideosm7la1g4jf2rghn0n89achgstehb.apps.googleusercontent.com"
  googleOauth2ClientSecret = "JaJjVGA5c6tSdluQdfFqNau8"
)

var gauthConfig goauth2.Config

func CreateGCECredential(rackName, projectId string) {
  listener, err := net.Listen("tcp", "127.0.0.1:0")
  checkErr(err)

  defer listener.Close()
  log.Printf("Oauth2 callback receiver listening on", listener.Addr().String())

  gauthConfig = goauth2.Config{
    Endpoint:     google.Endpoint,
    ClientID:     googleOauth2ClientID,
    ClientSecret: googleOauth2ClientSecret,
    Scopes:       []string{
      "https://www.googleapis.com/auth/cloudplatformprojects",
      "https://www.googleapis.com/auth/iam",
    },
    RedirectURL:  "http://" + listener.Addr().String(),
  }

  promptSelectAccount := goauth2.SetAuthURLParam("prompt", "select_account")
  codeURL := gauthConfig.AuthCodeURL("/", promptSelectAccount)

  log.Printf("Auhtorization code URL:", codeURL)
  open.Start(codeURL)

  http.Handle("/", callbackHandler{rackName, projectId, handleGauthCallback})
  http.Serve(listener, nil)
}

type callbackHandler struct {
  rackName , projectId string
  H func(*callbackHandler, http.ResponseWriter, *http.Request)
}

func(h callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
  h.H(&h, w, r)
  w.Write([]byte("DONE"))
  os.Exit(0)
}

// https://developers.google.com/identity/protocols/OAuth2InstalledApp
func handleGauthCallback(h *callbackHandler, w http.ResponseWriter, r *http.Request) {
  code := r.URL.Query().Get("code")
  if code == "" {
    return
  }

  err := EnsureApprc()
  checkErr(errors.Wrap(err, "failed to create config file"))

  token, err := gauthConfig.Exchange(context.Background(), code)
  checkErr(errors.Wrap(err, "invalid context"))

  client := goauth2.NewClient(context.Background(), goauth2.StaticTokenSource(&goauth2.Token{
    AccessToken: token.AccessToken,
  }))

  rmgrClient, err := crmgr.New(client)
  checkErr(errors.Wrap(err, "failed to get cloudsource manager"))

  presp, err := rmgrClient.Projects.List().Do()
  checkErr(errors.Wrap(err, "failed to list google projects"))

  if len(presp.Projects) == 0 {
    log.Fatal("No Google cloud project exists. Please create new Google Cloud project from web console: https://console.cloud.google.com")
  }

  var projectId string

  for _, p := range presp.Projects {
    if p.ProjectId ==  h.projectId || p.Name == h.rackName {
      projectId = p.ProjectId
      break
    }
  }

  if projectId == "" {
    projectId = helper.GenerateId(h.rackName, 4)
    p := &crmgr.Project{Name: h.rackName, ProjectId: projectId }
    _, err = rmgrClient.Projects.Create(p).Do()
    checkErr(errors.Wrap(err, fmt.Sprintf("failed to create project %+v", p)))
  }

  iamClient, err := iam.New(client)
  checkErr(errors.Wrap(err, "failed to create iam client"))

  saName := "razorctl"
  saFQN := fmt.Sprintf("projects/%v/serviceAccounts/%v@%v.iam.gserviceaccount.com", projectId, saName, projectId)
  _, err = iamClient.Projects.ServiceAccounts.Get(saFQN).Do()
  if err != nil {
    _, err = iamClient.Projects.ServiceAccounts.Create("projects/"+projectId, &iam.CreateServiceAccountRequest{
      AccountId: saName,
      ServiceAccount: &iam.ServiceAccount{
        DisplayName: "Razorbox cli svc account",
      },
    }).Do()

   checkErr(errors.Wrap(err, fmt.Sprintf("failed to create iam %q", saFQN)))
  }

  p, err := rmgrClient.Projects.GetIamPolicy(projectId, &crmgr.GetIamPolicyRequest{}).Do()
  checkErr(errors.Wrap(err, fmt.Sprintf("failed to get iam policy for %q", projectId)))

  members := []string{fmt.Sprintf("serviceAccount:%v@%v.iam.gserviceaccount.com", saName, projectId)}
  newPolicy :=  &crmgr.Policy{ 
      Bindings: []*crmgr.Binding{
        &crmgr.Binding{Role: "roles/viewer", Members: members},
        &crmgr.Binding{Role: "roles/deploymentmanager.editor", Members: members},
        &crmgr.Binding{Role: "roles/storage.admin", Members: members},
        &crmgr.Binding{Role: "roles/cloudbuild.builds.editor", Members: members},
      },
  }

  mergedBindings := mergeBindings(append(p.Bindings, newPolicy.Bindings...))
  mergedBindingsMap := rolesToMembersMap(mergedBindings)
  p.Bindings = rolesToMembersBinding(mergedBindingsMap)

  dump, _ := json.MarshalIndent(p.Bindings, " ", "  ")
  log.Printf(string(dump))

  _, err = rmgrClient.Projects.SetIamPolicy(projectId, &crmgr.SetIamPolicyRequest{Policy: p}).Do()
  checkErr(errors.Wrap(err, "failed to apply iam roles"))

  sKey, err := iamClient.Projects.ServiceAccounts.Keys.Create(saFQN, &iam.CreateServiceAccountKeyRequest{}).Do()
  checkErr(errors.Wrap(err, "failed to create iam key"))

  data, err := base64.StdEncoding.DecodeString(sKey.PrivateKeyData)
  checkErr(errors.Wrap(err, "failed to decode private key"))

  rc, err := LoadApprc()
  checkErr(err)
  
  auth := rc.GetAuth()
  if auth == nil { auth = &Auth{} }

  auth.RackName   = h.rackName
  auth.ProjectId  = projectId
  auth.ServiceKey = data

  err = SetAuth(auth)
  checkErr(errors.Wrap(err, "failed to save service key."))
}


// Convert a map of roles->members to a list of Binding
func rolesToMembersBinding(m map[string]map[string]bool) []*crmgr.Binding {
  bindings := make([]*crmgr.Binding, 0)
  for role, members := range m {
    b := crmgr.Binding{
      Role:    role,
      Members: make([]string, 0),
    }
    for m, _ := range members {
      b.Members = append(b.Members, m)
    }
    bindings = append(bindings, &b)
  }
  return bindings
}

func rolesToMembersMap(bindings []*crmgr.Binding) map[string]map[string]bool {
  bm := make(map[string]map[string]bool)
  // Get each binding
  for _, b := range bindings {
    // Initialize members map
    if _, ok := bm[b.Role]; !ok {
      bm[b.Role] = make(map[string]bool)
    }
    // Get each member (user/principal) for the binding
    for _, m := range b.Members {
      // Add the member
      bm[b.Role][m] = true
    }
  }
  return bm
}

func mergeBindings(bindings []*crmgr.Binding) []*crmgr.Binding {
  bm := rolesToMembersMap(bindings)
  rb := make([]*crmgr.Binding, 0)

  for role, members := range bm {
    var b crmgr.Binding
    b.Role = role
    b.Members = make([]string, 0)
    for m, _ := range members {
      b.Members = append(b.Members, m)
    }
    rb = append(rb, &b)
  }

  return rb
}
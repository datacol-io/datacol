package google

import (
  "fmt"
  "net"
  "net/http"
  "context"
  "encoding/base64"

  log "github.com/Sirupsen/logrus" 
  "github.com/skratchdot/open-golang/open"
  goauth2 "golang.org/x/oauth2"
  oauth2_google "golang.org/x/oauth2/google"
  crmgr "google.golang.org/api/cloudresourcemanager/v1"
  iam "google.golang.org/api/iam/v1"
)

const (
  googleOauth2ClientID     = "992213213700-ideosm7la1g4jf2rghn0n89achgstehb.apps.googleusercontent.com"
  googleOauth2ClientSecret = "JaJjVGA5c6tSdluQdfFqNau8"
  saName = "dcolctl"
  getUserEmail = true
)

var gauthConfig goauth2.Config
var projectId string
var pNumber int64

type authPacket struct {
  Cred []byte
  Err  error
  ProjectId string
  PNumber   int64
}

func CreateCredential(rackName, projectId string) authPacket {
  listener, err := net.Listen("tcp", "127.0.0.1:0")
  if err != nil {
    return authPacket{Err: err}
  }

  log.Debugf("Oauth2 callback receiver listening on %s", listener.Addr().String())

  scopes := []string{
    "https://www.googleapis.com/auth/iam",
    "https://www.googleapis.com/auth/cloudplatformprojects",
    "https://www.googleapis.com/auth/sqlservice.admin",
  }

  if getUserEmail {
    scopes = append(scopes, "https://www.googleapis.com/auth/userinfo.email")
  }

  gauthConfig = goauth2.Config{
    Endpoint:     oauth2_google.Endpoint,
    ClientID:     googleOauth2ClientID,
    ClientSecret: googleOauth2ClientSecret,
    Scopes:       scopes,
    RedirectURL:  "http://" + listener.Addr().String(),
  }

  promptSelectAccount := goauth2.SetAuthURLParam("prompt", "select_account")
  codeURL := gauthConfig.AuthCodeURL("/", promptSelectAccount)

  log.Debugf("Auhtorization code URL: %v", codeURL)
  open.Start(codeURL)

  stop := make(chan authPacket, 1)
  http.Handle("/", callbackHandler{rackName, projectId, handleGauthCallback, stop})
  
  go http.Serve(listener, nil)

  select {
    case msg := <- stop:
      listener.Close()
      return msg
  }

  return authPacket{Err: fmt.Errorf("create credentials")}
}

type callbackHandler struct {
  rackName , projectId string
  H func(*callbackHandler, http.ResponseWriter, *http.Request)([]byte, error)
  stop chan authPacket
}

func(h callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
  data, err := h.H(&h, w, r)

  if err != nil {
    h.termOnError(err)
    w.Write([]byte("Failed to authenticate. Please try again."))
    w.Write([]byte(fmt.Sprintf("\nError: %v", err)))
  } else {
    h.termOnSuccess(data)
    w.Write([]byte("Successfully authenticated. Please go to your terminal."))
  }
}

func (h callbackHandler) termOnError(err error) {
  h.stop <- authPacket{Err: err}
}

func (h callbackHandler) termOnSuccess(data []byte){
  h.stop <- authPacket{Cred: data, ProjectId: projectId, PNumber: pNumber}
}

// https://developers.google.com/identity/protocols/OAuth2InstalledApp
func handleGauthCallback(h *callbackHandler, w http.ResponseWriter, r *http.Request) ([]byte, error) {
  code := r.URL.Query().Get("code")
  var cred []byte

  if code == "" {
    return cred, fmt.Errorf("invalid code")
  }

  token, err := gauthConfig.Exchange(context.Background(), code)
  if err != nil {
    return cred, fmt.Errorf("invalid context: %v", err)
  }

  client := goauth2.NewClient(context.Background(), goauth2.StaticTokenSource(&goauth2.Token{
    AccessToken: token.AccessToken,
  }))

  rmgrClient, err := crmgr.New(client)
  if err != nil {
    return cred, fmt.Errorf("failed to get cloudsource manager: %v", err)
  }

  presp, err := rmgrClient.Projects.List().Do()
  if err != nil {
    return cred, fmt.Errorf("failed to list google projects")
  }

  if len(presp.Projects) == 0 {
    return cred, fmt.Errorf("No Google cloud project exists. Please create new Google Cloud project from web console: https://console.cloud.google.com")
  }

  projectId = ""
  for _, p := range presp.Projects {
    if p.ProjectId ==  h.projectId || p.Name == h.projectId {
      projectId = p.ProjectId
      pNumber   = p.ProjectNumber
      break
    }
  }

  if projectId == "" {
    return cred, fmt.Errorf("failed to get project %v", h.projectId)
  }

  log.Debugf("Selected ProjectId: %s", projectId)

  if getUserEmail {
    if err := addToContactList(token.AccessToken); err != nil {
      log.WithFields(log.Fields{"project": projectId}).Debugf(err.Error())
    }
  }

  iamClient, err := iam.New(client)
  if err != nil {
    return cred, fmt.Errorf("failed to create iam client: %v", err)
  }

  saFQN := fmt.Sprintf("projects/%v/serviceAccounts/%v@%v.iam.gserviceaccount.com", projectId, saName, projectId)
  _, err = iamClient.Projects.ServiceAccounts.Get(saFQN).Do()
  if err != nil {
    _, err = iamClient.Projects.ServiceAccounts.Create("projects/"+projectId, &iam.CreateServiceAccountRequest{
      AccountId: saName,
      ServiceAccount: &iam.ServiceAccount{
        DisplayName: "Datacol cli service account",
      },
    }).Do()

    if err != nil {
     return cred, fmt.Errorf("failed to create iam %q", saFQN)
    }
  }

  p, err := rmgrClient.Projects.GetIamPolicy(projectId, &crmgr.GetIamPolicyRequest{}).Do()
  if err != nil {
    return cred, fmt.Errorf("failed to get iam policy for %q", projectId)
  }

  members := []string{fmt.Sprintf("serviceAccount:%v@%v.iam.gserviceaccount.com", saName, projectId)}
  newPolicy :=  &crmgr.Policy{ 
      Bindings: []*crmgr.Binding{
        &crmgr.Binding{Role: "roles/viewer", Members: members},
        &crmgr.Binding{Role: "roles/deploymentmanager.editor", Members: members},
        &crmgr.Binding{Role: "roles/storage.admin", Members: members},
        &crmgr.Binding{Role: "roles/cloudbuild.builds.editor", Members: members},
        &crmgr.Binding{Role: "roles/container.developer", Members: members},
        &crmgr.Binding{Role: "roles/cloudsql.admin", Members: members},
        &crmgr.Binding{Role: "roles/cloudsql.client", Members: members},
      },
  }

  mergedBindings := mergeBindings(append(p.Bindings, newPolicy.Bindings...))
  mergedBindingsMap := rolesToMembersMap(mergedBindings)
  p.Bindings = rolesToMembersBinding(mergedBindingsMap)

  log.Debugf("Creating IAM permissions: +%v \n", string(toJson(p.Bindings)))

  _, err = rmgrClient.Projects.SetIamPolicy(projectId, &crmgr.SetIamPolicyRequest{Policy: p}).Do()
  if err != nil {
    return cred, fmt.Errorf("failed to apply IAM roles")
  }

  sKey, err := iamClient.Projects.ServiceAccounts.Keys.Create(saFQN, &iam.CreateServiceAccountKeyRequest{}).Do()
  if err != nil {
    return cred, fmt.Errorf("failed to create iam key: %v", err)
  }

  cred, err = base64.StdEncoding.DecodeString(sKey.PrivateKeyData)
  if err != nil {
    return cred, fmt.Errorf("failed to decode private key: %v", err)
  }

  return cred, nil
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

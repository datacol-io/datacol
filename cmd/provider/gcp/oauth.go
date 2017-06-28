package gcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go-term"
	"github.com/skratchdot/open-golang/open"
	goauth2 "golang.org/x/oauth2"
	oauth2_google "golang.org/x/oauth2/google"
	crmgr "google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
)

const (
	googleOauth2ClientID     = "992213213700-ideosm7la1g4jf2rghn0n89achgstehb.apps.googleusercontent.com"
	googleOauth2ClientSecret = "JaJjVGA5c6tSdluQdfFqNau8"
	saPrefix                 = "datacol"
	welcomeMessage           = `Welcome to Datacol CLI. This command will guide you through creating a new infrastructure inside your Google account. 
It uses various Google services (like Container engine, Container builder, Deployment Manager etc) under the hood to 
automate all away to give you a better deployment experience.

Datacol CLI will authenticate with your Google Account and install the Datacol platform into your GCP account. 
These credentials will only be used to communicate between this installer running on your computer and the Google platform.`
)

var (
	gauthConfig goauth2.Config
	projectId   string
	pNumber     int64
	emailOptin  = true
)

type authPacket struct {
	Cred        []byte
	Err         error
	ProjectId   string
	AccessToken string
	PNumber     int64
	SAEmail     string
}

func CreateCredential(rackName string, optout bool) authPacket {
	fmt.Printf(welcomeMessage)
	prompt("")

	emailOptin = !optout

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return authPacket{Err: err}
	}

	log.Debugf("Oauth2 callback receiver listening on %s", listener.Addr().String())

	scopes := []string{
		"https://www.googleapis.com/auth/iam",
		"https://www.googleapis.com/auth/cloud-platform",
	}

	if emailOptin {
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

	if err := open.Start(codeURL); err != nil {
		term.ExitOnError(err)
	}

	stop := make(chan authPacket, 1)
	http.Handle("/", callbackHandler{rackName, handleGauthCallback, stop})

	go http.Serve(listener, nil)

	select {
	case msg := <-stop:
		listener.Close()
		client, err := setIamPolicy(rackName, msg.AccessToken)
		if err != nil {
			return authPacket{Err: err}
		}

		email, data, err := NewServiceAccountPrivateKey(client, rackName, projectId)
		if err != nil {
			return authPacket{Err: err}
		}

		msg.Cred = data
		msg.ProjectId = projectId
		msg.SAEmail = email
		return msg
	}
}

type callbackHandler struct {
	rackName string
	H        func(*callbackHandler, http.ResponseWriter, *http.Request) (string, error)
	stop     chan authPacket
}

func (h callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := h.H(&h, w, r)

	if err != nil {
		h.termOnError(err)
		w.Write([]byte("Failed to authenticate. Please try again."))
		w.Write([]byte(fmt.Sprintf("\nError: %v", err)))
	} else {
		h.termOnSuccess(token)
		w.Write([]byte("Successfully authenticated. Please go to your terminal."))
	}
}

func (h callbackHandler) termOnError(err error) {
	h.stop <- authPacket{Err: err}
}

func (h callbackHandler) termOnSuccess(token string) {
	h.stop <- authPacket{
		AccessToken: token,
		ProjectId:   projectId,
		PNumber:     pNumber,
	}
}

// https://developers.google.com/identity/protocols/OAuth2InstalledApp
func handleGauthCallback(h *callbackHandler, w http.ResponseWriter, r *http.Request) (string, error) {
	code := r.URL.Query().Get("code")

	if code == "" {
		return "", fmt.Errorf("invalid code")
	}

	token, err := gauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return "", fmt.Errorf("invalid context: %v", err)
	}

	if emailOptin {
		go subscribeMe(token.AccessToken)
	}

	return token.AccessToken, nil
}

func setIamPolicy(rackName, accessToken string) (*iam.Service, error) {
	client := goauth2.NewClient(context.Background(), goauth2.StaticTokenSource(&goauth2.Token{
		AccessToken: accessToken,
	}))

	rmgrClient, err := crmgr.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloudsource-manager: %v", err)
	}

	presp, err := rmgrClient.Projects.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list google projects")
	}

	if len(presp.Projects) == 0 {
		return nil, fmt.Errorf("No Google cloud project exists. Please create new Google Cloud project from web console: https://console.cloud.google.com and retry.")
	}

	projects := make([]string, len(presp.Projects))
	for i, p := range presp.Projects {
		projects[i] = p.Name
	}

	fmt.Println("\nPlease choose a project to continue.")

	i, _ := term.List(projects)
	projectId = presp.Projects[i].ProjectId
	pNumber = presp.Projects[i].ProjectNumber

	if len(projectId) == 0 {
		return nil, fmt.Errorf("Please select atleast an option.")
	}

	log.Debugf("Selected ProjectId: %s", projectId)

	iamClient, err := iam.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create iam client: %v", err)
	}

	saName := fmt.Sprintf("%s-%s", saPrefix, rackName)
	svcName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, projectId)
	saFQN := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectId, svcName)

	_, err = iamClient.Projects.ServiceAccounts.Get(saFQN).Do()
	if err != nil {
		_, err = iamClient.Projects.ServiceAccounts.Create("projects/"+projectId, &iam.CreateServiceAccountRequest{
			AccountId: saName,
			ServiceAccount: &iam.ServiceAccount{
				DisplayName: fmt.Sprintf("Service account for %s stack created by Datacol", rackName),
			},
		}).Do()

		if err != nil {
			return nil, fmt.Errorf("failed to create iam service account %q err: %v", saFQN, err)
		}
	}

	p, err := rmgrClient.Projects.GetIamPolicy(projectId, &crmgr.GetIamPolicyRequest{}).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get iam policy for %q", projectId)
	}

	members := []string{fmt.Sprintf("serviceAccount:%s", svcName)}
	newPolicy := &crmgr.Policy{
		Bindings: []*crmgr.Binding{
			&crmgr.Binding{Role: "roles/owner", Members: members},
		},
	}

	mergedBindings := mergeBindings(append(p.Bindings, newPolicy.Bindings...))
	mergedBindingsMap := rolesToMembersMap(mergedBindings)
	p.Bindings = rolesToMembersBinding(mergedBindingsMap)

	log.Debugf("Creating IAM permissions: +%v \n", string(toJson(p.Bindings)))

	_, err = rmgrClient.Projects.SetIamPolicy(projectId, &crmgr.SetIamPolicyRequest{Policy: p}).Do()
	if err != nil {
		log.Warn(err)
		return nil, fmt.Errorf("failed to apply IAM roles")
	}

	return iamClient, nil
}

func NewServiceAccountPrivateKey(iamClient *iam.Service, rackName, pid string) (string, []byte, error) {
	cred := []byte{}
	if len(projectId) == 0 {
		projectId = pid
	}

	saName := fmt.Sprintf("%s-%s", saPrefix, rackName)
	svcName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, projectId)
	saFQN := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectId, svcName)

	log.Debugf("creating new private key for %s", saFQN)

	sKey, err := iamClient.Projects.ServiceAccounts.Keys.Create(saFQN, &iam.CreateServiceAccountKeyRequest{}).Do()
	if err != nil {
		return svcName, cred, fmt.Errorf("failed to create iam key: %v", err)
	}

	cred, err = base64.StdEncoding.DecodeString(sKey.PrivateKeyData)
	if err != nil {
		return svcName, cred, fmt.Errorf("failed to decode private key: %v", err)
	}

	return svcName, cred, nil
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

func subscribeMe(accessToken string) {
	if err := addToContactList(accessToken); err != nil {
		log.WithFields(log.Fields{"project": projectId}).Debugf(err.Error())
	}
}

package google

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "io/ioutil"
  log "github.com/Sirupsen/logrus"
)

var (
  mcAPIKey string
)

func init() {
  // mcAPIKey = os.Getenv("MAILCHIMP_API_KEY")
  mcAPIKey = "b9e72c3d6982efb7645dca08c57de53e-us5"
}

type payloadMCSignup struct {
  Email       string `json:"email_address"`
  Status      string `json:"status"`
  StatusIfNew string `json:"status_if_new"`
}

type gprofile struct {
  Name   string  `json: "name"`
  Email  string  `json: "email"`
}


// Signup for Mailchimp newsletter
func signup(pf *gprofile) error {
  mcAPIHost := "https://us5.api.mailchimp.com/3.0"
  mcListID  := "619a22cb36"

  payload := payloadMCSignup{
    Email:       pf.Email,
    Status:      "subscribed",
    StatusIfNew: "subscribed",
  }

  jsonBytes, _ := json.Marshal(payload)
  requestPath := fmt.Sprintf("%s/lists/%s/members", mcAPIHost, mcListID)

  request, err := http.NewRequest("POST", requestPath, bytes.NewBuffer(jsonBytes))
  request.SetBasicAuth("username", mcAPIKey)

  client := &http.Client{}
  response, err := client.Do(request)
  if err != nil {
    return err
  }
  defer response.Body.Close()

  body, _ := ioutil.ReadAll(response.Body)
  var raw interface{}
  if err = json.Unmarshal(body, &raw); err != nil {
    return err
  }

  if response.StatusCode != http.StatusOK {
    return fmt.Errorf("invalid response code:%d body: %+v", response.StatusCode, raw)
  }

  return nil
}

func addToContactList(token string) error {
  requestPath := "https://www.googleapis.com/oauth2/v2/userinfo"
  request, err := http.NewRequest("GET", requestPath, nil)
  if err != nil { return err }

  request.Header.Set("Authorization", "Bearer " + token)

  client := &http.Client{}
  response, err := client.Do(request)
  if err != nil { return err }

  defer response.Body.Close()

  if response.StatusCode != http.StatusOK {
    return fmt.Errorf("Invalid response: %v", response)
  } else {
    body, _ := ioutil.ReadAll(response.Body)
    var gp gprofile
    err := json.Unmarshal(body, &gp)
    if err != nil {
      return err 
    }
    
    log.Debugf("user profile: %+v", gp)
    return signup(&gp)
  }
  return nil
}


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
  apiHost string
)

func init() {
  apiHost = "http://api.datacol.io"
  apiHost = "https://api-ahlkzhcseh.now.sh"
}

type gprofile struct {
  Name   string  `json: "name"`
  Email  string  `json: "email"`
}

// subscribe for datacol newsletter
func subscribe(pf *gprofile) error {
  jsonBytes, _ := json.Marshal(pf)
  requestPath := fmt.Sprintf("%s/subscribe", apiHost)
  request, err := http.NewRequest("POST", requestPath, bytes.NewBuffer(jsonBytes))

  client := &http.Client{}
  response, err := client.Do(request)
  if err != nil {
    return err
  }
  defer response.Body.Close()

  if response.StatusCode != http.StatusOK {
    return fmt.Errorf("invalid response code:%d", response.StatusCode)
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
    return subscribe(&gp)
  }
  return nil
}


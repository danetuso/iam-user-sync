package main

import (
  "fmt"
  "io/ioutil"

  "golang.org/x/net/context"
  "golang.org/x/oauth2/google"

  admin "google.golang.org/api/admin/directory/v1"
  "google.golang.org/api/option"
  "encoding/json"
  "strings"
)


// RsaKey struct to map json RawMessage to
type RsaKey struct {
	Key string `json:"Public_SSH_Key"`
}

// CreateDirectoryService builds and returns an Admin SDK Directory service object authorized with
// the service accounts that act on behalf of the given user.
func CreateDirectoryService(userEmail string, credentialsPath string) (*admin.Service, error) {
  ctx := context.Background()

  jsonCredentials, err := ioutil.ReadFile(credentialsPath)
  if err != nil {
    return nil, err
  }

  config, err := google.JWTConfigFromJSON(jsonCredentials, admin.AdminDirectoryUserScope)
  if err != nil {
    return nil, fmt.Errorf("JWTConfigFromJSON: %v", err)
  }
  config.Subject = userEmail

  ts := config.TokenSource(ctx)

  srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
  if err != nil {
    return nil, fmt.Errorf("NewService: %v", err)
  }
  return srv, nil
}

// PullGsuiteUsers Creates a Google Workspace API Directory Service and authenticates. Queries the API for a list of domain users with the appropriate custom attribute set.
// Returns a list of gsuiteUser objects.
func PullGsuiteUsers(email string, domain string, mask string, credentialsPath string) ([]IAMUser, error) {

	// gsuiteUsers List of gsuiteUser objects
	var gsuiteUsers = []IAMUser {}

	srv, e := CreateDirectoryService(email, credentialsPath)
	if e != nil {
		return nil, e
	}
	r, err := srv.Users.List().Domain(domain).Projection("Custom").CustomFieldMask(mask).Do()
	if err != nil {
		return nil, err
	}

	if len(r.Users) != 0 {
		for _, u := range r.Users {
				if val, ok := u.CustomSchemas["SSHKEY"]; ok {
					// Custom Schema SSHKEY exists

					var rsakey RsaKey
					err := json.Unmarshal(val, &rsakey)
					if err != nil {
						return nil, err
					}

					uName := u.Name.GivenName + "." + u.Name.FamilyName
					gUser := IAMUser{
						username: strings.ToLower(uName),
						publickey: rsakey.Key,
					}
					gsuiteUsers = append(gsuiteUsers, gUser)
				}
		}
	}
	return gsuiteUsers, nil
}
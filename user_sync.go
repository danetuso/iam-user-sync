package main

import (
	"os"
	"os/user"
	"os/exec"
	"strings"
	"io/ioutil"
	"strconv"
	"flag"
	"gopkg.in/yaml.v2"
	"log"
	"errors"
	"time"
)

// IAMUser struct for maintaining local users
type IAMUser struct {
	username string
	publickey string
}

// Config struct defines parameters where user input is necessary
type Config struct {
    Provider string `yaml:"provider"`
	Credentials string `yaml:"credentials"`
	Group string `yaml:"group"`
	KeepHomeDir bool `yaml:"keephomedir"`
	LogFile string `yaml:"logfile"`
	Email string `yaml:"gsuiteadmin"`
	Domain string `yaml:"oauthdomain"`
	CustomAttributeKey string `yaml:"customattributekey"`
}

// Cfg Globally accessed Config struct
var Cfg Config

// globalLogger is a globally accessed pointer to the custom logger
var globalLogger *Logger

func main() {
	start := time.Now()
	
	// Process Arguments
	argErr := ProcessInput()
	if argErr != nil {
		log.Printf("Fatal Error! Problem processing arguments: %v\n", argErr)
		return
	}
	
	// Initialize Logging
    globalLogger, err := NewFileLogger(Cfg.LogFile)
    if err != nil {
        log.Printf("Error while initializing logging: %v\n", err)
    }

	globalLogger.Info("====== Start Log ======\n")
	globalLogger.Info("IAM User Sync starting with configuration settings: Provider: %s | Group: %s | KeepHomeDir: %t | LogFile: %s\n", Cfg.Provider, Cfg.Group, Cfg.KeepHomeDir, Cfg.LogFile)
	if Cfg.Provider == "GSUITE" {
		globalLogger.Info("GSUITE configuration settings: Email: %s | Domain: %s | Custom Attribute Key: %s | Path To Credentials: %s\n", Cfg.Email, Cfg.Domain, Cfg.CustomAttributeKey, Cfg.Credentials)
	}


	// If the group does not exist, create a group then continue
	if !doesGroupExist(Cfg.Group) {
		log.Printf("Group %s not found on local system.\n", Cfg.Group)
		createGroupError := createGroup(Cfg.Group)
		if createGroupError != nil {
			globalLogger.Error("Problem creating group: %v\n", createGroupError)
			return
		} else {
			globalLogger.Info("Group %s created successfully.\n", Cfg.Group)
		}
	}
	

	// =======================
	// CHECK FOR USERS TO ADD
	// =======================

	// Define the list of user structs from IAM
	var users = []IAMUser {}
	var pullUsersError error
	users, pullUsersError = PullUsersFromIAM()
	if pullUsersError != nil {
		globalLogger.Error("Issue pulling users from IAM: %v\n", pullUsersError)
		return
	}
	if users == nil {
		globalLogger.Error("Provider %s not supported!\n", Cfg.Provider)
		return
	}
	if len(users) < 1 {
		globalLogger.Error("List of IAM Users is empty!\n")
		return
	}

	// Define and pull the list of local users
	localUsersList, localUserError := getUsersInGroup(Cfg.Group)
	if localUserError != nil {
		globalLogger.Error("Issue pulling list of local users in group with error: %v\n", localUserError)
		return
	}

	// Define the list of iam usernames
	iamUsersList := []string {}

	// Iterate through iam users list of structs
    for _, usr := range users {
		// Add the users from the struct to a list for comparison later
		iamUsersList = append(iamUsersList, usr.username)

		// Check to see if the iam username already has a local user on this machine
		localUserExists := false
		for _, localUser := range localUsersList {
			if localUser == usr.username {
				//Local user found, do not add
				localUserExists = true
			}
		}

		// IAM user doesn't exist locally, so create a new user
		if !localUserExists {
			addUserError := addUser(usr)
			if addUserError != nil {
				globalLogger.Error("Issue creating local user %s with error: %v\n", usr.username, addUserError)
				return
			} else {
				globalLogger.Info("New user found in IAM that does not exist locally! Adding user: %s\n", usr.username)
			}
		} else {
			// User already exists, but ensure the authorized_keys file exists
			createAuthorizedKeysError := createAuthorizedKeys(usr)
			if createAuthorizedKeysError != nil {
				globalLogger.Error("Problem creating .ssh/authorized_keys file for pre-existing user with error: %v\n", createAuthorizedKeysError)
				return
			}
		}
    }

	// =======================
	// CHECK FOR USERS TO DELETE
	// =======================

	// Repopulate the local users list in case any were added
	localUsersList, localUserError = getUsersInGroup(Cfg.Group)
	if localUserError != nil {
		globalLogger.Error("Issue pulling list of local users in group with error: %v\n", localUserError)
		return
	}
	
	// Check to see if there are any users locally that aren't in the iam users, if so then delete user (keep home folder?)
	if len(localUsersList) > 0 {

		// Iterate through local users list
		for _, localUser := range localUsersList {
		
			// Check to see if the local username is in the list of iam usernames populated above
			iamUserExists := false
			for _, iamUserName := range iamUsersList {
				if localUser == iamUserName {
					// Local user is in iam user list as intended, proceed
					iamUserExists = true
				}
			}
	
			// Local user does not match a record in iam, delete the local user
			if !iamUserExists {
				globalLogger.Info("Stale user found! Deleting user: %s\n", localUser)
				if !Cfg.KeepHomeDir {
					globalLogger.Info(localUser + "'s home folder has been deleted!\n")
				}
				deleteUserError := deleteUser(localUser, Cfg.KeepHomeDir)
				if deleteUserError != nil {
					globalLogger.Error("Issue deleting user: %v\n", deleteUserError)
					return
				}
				
			}
		}
	}


	duration := time.Since(start)
	globalLogger.Info("====== End Log (Done in %dms) ======\n", duration.Milliseconds())

	// Close Logging
	closeLogErr := globalLogger.CloseFile()
	if closeLogErr != nil {
		log.Printf("Fatal Error! Error while closing the log file: %v\n", closeLogErr)
		return
	}
}

// ProcessInput creates a Config object using supplied parameters and/or variables defined in config.yml
func ProcessInput() error {
	provider := flag.String("provider", "", "Available Choices: GSUITE, AWS, AZURE (Default: GSUITE)")
	credentials := flag.String("credentials", "", "Path to IAM Provider service account credentials file. (Default: ./credentials.json)")
	group := flag.String("group", "", "Name of the group to be maintained. (Default: iamusersync)")
	keepHomeDir := flag.Bool("keephomedir", false, "Decides whether to delete a user's home folder when being removed.")
	logFile := flag.String("logfile", "", "Path to file that you want to log output to. (Default: /var/log/iamusersync.log)")

	gsuiteAdminEmail := flag.String("gsuiteadmin", "", "GSuite admin email that delegated OAuth scopes. If the IAM Provider is GSUITE, this is required.") 
	gsuiteOAuthDomain := flag.String("oauthdomain", "", "GSuite OAuth domain, only if different from Admin's email domain.")
	customAttributeKey := flag.String("customattributekey", "", "Gsuite user custom attribute key name. See README for more details. (Default: SSHKEY)")
	
	config := flag.String("config", "", "Full path to config file. Additional arguments supplied on the CLI will override parameters set in config.")

	flag.Parse()

	// if config path is set, use those values
	if *config != "" {
		f, err := os.Open(*config)
		if err != nil {
			return err
		}
		defer f.Close()
		
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&Cfg)
		if err != nil {
			return err
		}
		log.Printf("Loading config from: %s\n", *config)
	}

	// if cli parameter is passed, overwite config variables
	// general config:
	if *group != "" {
		Cfg.Group = *group
	}
	if *keepHomeDir != false {
		Cfg.KeepHomeDir = *keepHomeDir
	}
	if *logFile != "" {
		Cfg.LogFile = *logFile
	}
	if *provider != "" {
		Cfg.Provider = *provider
	}

	// gsuite config:
	if *credentials != "" {
		Cfg.Credentials = *credentials
	}
	if *gsuiteAdminEmail != "" {
		Cfg.Email = *gsuiteAdminEmail
	}
	if *gsuiteOAuthDomain != "" {
		Cfg.Domain = *gsuiteOAuthDomain
	}
	if *customAttributeKey != "" {
		Cfg.CustomAttributeKey = *customAttributeKey
	}

	// if IAM Provider is GSUITE, gsuiteadminemail must be set either from parameter or config
	if (*provider == "GSUITE" && *gsuiteAdminEmail == "") || (Cfg.Provider == "GSUITE" && Cfg.Email == "") {
		emailMissingError := errors.New("If the IAM provider is GSuite (Google Workspace) then you must supply the super admin user's email address that delegated the serviceaccount OAuth scopes.")
		return emailMissingError
	}
	
	// if required parameters are empty, error
	if Cfg.Provider == "" {
		providerMissingError := errors.New("IAM Provider must be set. Please use --provider GSUITE or set the value using --config /path/to/config.yml.")
		return providerMissingError
	}
	if Cfg.Credentials == "" {
		credentialsMissingError := errors.New("IAM Provider service account credentials must be present. Please use --credentials ./credentials.json or set the value in config.yml.")
		return credentialsMissingError
	}

	// Use defaults where arguments not specified
	if Cfg.Group == "" {
		Cfg.Group = "iamusersync"
		log.Printf("Group not specified. Using default: %s\n", Cfg.Group)
	}
	if Cfg.LogFile == "" {
		Cfg.LogFile = "/var/log/iamusersync.log"
		log.Printf("Log file path not specified. Using default: %s\n", Cfg.LogFile)
	}
	if Cfg.Domain == "" {
		Cfg.Domain = strings.Split(Cfg.Email, "@")[1]
		log.Printf("Using domain for user lookup: %s\n", Cfg.Domain)
	}
	if Cfg.CustomAttributeKey == "" {
		Cfg.CustomAttributeKey = "SSHKEY"
		log.Printf("Gsuite User Custom Attribute Key not specified. Using default: %s\n", Cfg.CustomAttributeKey)
	}

	return nil
}

// addUser adds the given gsuiteUser to the local system using the useradd command.
// It then adds that user to the proper group and generates ~/.ssh/authorized_keys.
func addUser(u IAMUser) error {
	command := "useradd"
	param1 := "-m"
	param2 := "-d"
	param3 := "/home/" + u.username
	param4 := u.username
	cmd := exec.Command(command, param1, param2, param3, param4)
	_, err := cmd.Output()

    if err != nil {
		return err
    }

	addUserError := addUserToGroup(Cfg.Group, u.username)
	if addUserError != nil {
		return addUserError
	}

	createAuthorizedKeysError := createAuthorizedKeys(u)
	if createAuthorizedKeysError != nil {
		return createAuthorizedKeysError
	}

	return nil
}

// createAuthorizedKeys checks to see if a users home folder, .ssh folder, and authorized_keys file exist, 
// and if not then creates them with appropriate permissions and the public key pulled from IAM.
func createAuthorizedKeys(u IAMUser) error {
	homePath := "/home/" + u.username
	sshPath := homePath + "/.ssh/"
	authorizedKeysPath := sshPath + "authorized_keys"

	// check home path exists, if not then create it
	_, homeStatErr := os.Stat(homePath)
	if errors.Is(homeStatErr, os.ErrNotExist) {
		// user home does not exist
		mkdirErr := os.Mkdir(homePath, 0644)
		if mkdirErr != nil {
			return mkdirErr
		}

		chownErr := chown(u.username, homePath)
		if chownErr != nil {
			return chownErr
		}
		globalLogger.Info("Home folder not detected for user. Populating %s\n", authorizedKeysPath)
	}

	// check .ssh path exists, if not then create it
	_, sshStatErr := os.Stat(sshPath)
	if errors.Is(sshStatErr, os.ErrNotExist) {
		// user home does not exist
		mkdirErr := os.Mkdir(sshPath, 0644)
		if mkdirErr != nil {
			return mkdirErr
		}

		chownErr := chown(u.username, sshPath)
		if chownErr != nil {
			return chownErr
		}
	}

	// check authorized_keys file exists, if not then create it
	_, akStatErr := os.Stat(authorizedKeysPath)
	if errors.Is(akStatErr, os.ErrNotExist) {
		// authorized_keys file does not exist
		publicKeyWriteErr := ioutil.WriteFile(authorizedKeysPath, []byte(u.publickey), 0644)
		if publicKeyWriteErr != nil {
			return publicKeyWriteErr
		}

		chownErr := chown(u.username, authorizedKeysPath)
		if chownErr != nil {
			return chownErr
		}
	}

	return nil
}

// deleteUser removes a given username from the local system using the deluser command.
// If keepHomeDir is set to false, the given user's home directory will be deleted.
func deleteUser(username string, keepHomeDir bool) error {
	command := "deluser"
	param1 := username
	out := exec.Command(command, param1)
	_, err := out.Output()
	if err != nil {
		return err
	}

	if !keepHomeDir {
		cmd := "rm"
		p1 := "-rf"
		p2 := "/home/" + username
		delout := exec.Command(cmd, p1, p2)
		_, e := delout.Output()
		if e != nil {
			return e
		}
	}
	return nil
}

// doesGroupExist checks if a given group name exists on the local system, returns a boolean.
func doesGroupExist(name string) bool {
	_, err := user.LookupGroup(name)
	if err != nil {
		if _, ok := err.(user.UnknownGroupError); ok {
			return false
		}
	}
	return true
}

// createGroup creates a local system group with the given localGroupName using the groupadd command.
func createGroup(localGroupName string) error {
	command := "groupadd"
	param1 := localGroupName
	cmd := exec.Command(command, param1)
	_, err := cmd.Output()

    if err != nil {
        return err
    }
	return nil
}

// getUsersInGroup returns a list of username strings for a given group.
func getUsersInGroup(group string) ([]string, error) {
	command := "getent"
	param1 := "group"
	param2 := group
	cmd := exec.Command(command, param1, param2)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	groupString := string(stdout)

	concatUsers := strings.Split(groupString, ":")[3]
	trimmedConcatUsers := strings.TrimSpace(concatUsers)
	localUsers := strings.Split(trimmedConcatUsers, ",")
	return localUsers, nil
}

// addUserToGroup adds a given username to a given group.
func addUserToGroup(group string, username string) error {
	command := "usermod"
	param1 := "-aG"
	param2 := group
	param3 := username
	cmd := exec.Command(command, param1, param2, param3)
	_, err := cmd.Output()
	return err
}

// chown makes the given username own the given path on the local system.
func chown(username string, path string) error { 
	userObject, _ := user.Lookup(username)
	uid, _ := strconv.Atoi(userObject.Uid)
	gid, _ := strconv.Atoi(userObject.Gid)
	err := os.Chown(path, uid, gid)
	return err
}

// PullUsersFromIAM determines which class to use given a configured provider
func PullUsersFromIAM() ([]IAMUser, error) {
	p := strings.ToUpper(Cfg.Provider)
	switch p {
	case "GSUITE":
		return PullGsuiteUsers(Cfg.Email, Cfg.Domain, Cfg.CustomAttributeKey)
	case "AWS":
		return nil, nil
	// additional providers coming soon
	default:
		return nil, nil
	}
	return nil, nil
}

package library

import (
	log "../../core/logger"
	"../../core/arguments"
	"../../core/methods"
	"../objects"
	"strconv"
	"io/ioutil"
	"strings"
	"fmt"
	"os"
	"os/exec"
)

// Create environment file of this installation
func CreateEnvFile(t string) error {

	// Environment file fully qualified path
	objects.EnvFileName = arguments.EnvFileDir + "env_" + arguments.RequestedInstallVersion + "_" + t
	log.Println("Creating environment file for this installation at: " + objects.EnvFileName)

	// Create the file
	err := methods.CreateFile(objects.EnvFileName)
	if err != nil { return err }

	// Build arguments to write
	var EnvFileContents []string
	EnvFileContents = append(EnvFileContents, "source " + objects.BinaryInstallLocation + "/greenplum_path.sh")
	EnvFileContents = append(EnvFileContents, "export MASTER_DATA_DIRECTORY=" + objects.GpInitSystemConfig.MasterDir + "/" + objects.GpInitSystemConfig.ArrayName + "-1")
	EnvFileContents = append(EnvFileContents, "export PGPORT=" + strconv.Itoa(objects.GpInitSystemConfig.MasterPort))
	EnvFileContents = append(EnvFileContents, "export PGDATABASE=" + objects.GpInitSystemConfig.DBName)

	// Write to EnvFile
	err = methods.WriteFile(objects.EnvFileName, EnvFileContents)
	if err != nil { return err }

	return nil
}

// Check if there is any previous installation of the same version
func PrevEnvFile(product string) (string, error) {

	log.Println("Checking if there is previous installation for the version: " + arguments.RequestedInstallVersion)
	var MatchingFilesInDir []string
	allfiles, err := ioutil.ReadDir(arguments.EnvFileDir)
	if err != nil { return "", err }
	for _, file := range allfiles {

		if strings.Contains(file.Name(), arguments.RequestedInstallVersion) {
			MatchingFilesInDir = append(MatchingFilesInDir, file.Name())
		}

	}

	// Found matching environment file of this installation, now ask for confirmation
	if len(MatchingFilesInDir) > 1 {

		// Show all the environment files
		log.Warn("Found matching environment file for the version: " + arguments.RequestedInstallVersion)
		log.Println("Below are the list of environment file of the version: " + arguments.RequestedInstallVersion + "\n")

		// Temp files
		temp_env_file := arguments.TempDir + "temp_env.sh"
		temp_env_out_file := arguments.TempDir + "temp_env.out"

		// Create those files
		_ = methods.DeleteFile(temp_env_file)
		err := methods.CreateFile(temp_env_file)
		if err != nil { return "", err }

		// Bash script
		var cmd []string
		bashCmd :=      "incrementor=1" +
				";echo -e \"ID\t\t\tEnvironment File\t\t\tMaster Port\t\t\tStatus\"   > " + temp_env_out_file +
				";echo \"------------------------------------------------------------------------------------------------------------------------\"    >> " + temp_env_out_file +
				";ls -1 " + arguments.EnvFileDir + " | grep env_"+ arguments.RequestedInstallVersion +" | while read line" +
				";do    " +
				"       source "+arguments.EnvFileDir+"/$line" +
				"       ;psql -d template1 -p $PGPORT -Atc \"select 1\" &>/dev/null" +
				"       ;retcode=$?" +
				"       ;if [ \"$retcode\" == \"0\" ]; then" +
				"               echo -e \"$incrementor\t\t\t$line\t\t\t$PGPORT\t\t\tRUNNING\" >> " + temp_env_out_file +
				"       ;else" +
				"               echo -e \"$incrementor\t\t\t$line\t\t\t$PGPORT\t\t\tUNKNOWN/STOPPED/FAILED\"  >> " + temp_env_out_file +
				"       ;fi" +
				"       ;incrementor=$((incrementor+1))" +
				";done"
		cmd = append(cmd, bashCmd)

		// Copy it to the file
		_ = methods.WriteFile(temp_env_file, cmd)

		// Execute the script
		_, err = exec.Command("/bin/sh", temp_env_file).Output()
		if err != nil { return "", err }

		// Display the output
		out, _ := ioutil.ReadFile(temp_env_out_file)
		fmt.Println(string(out))

		// Cleanup the temp files
		_ = methods.DeleteFile(temp_env_file)
		_ = methods.DeleteFile(temp_env_out_file)

		// Create a list of the options
		var envStore []string
		for _, e := range MatchingFilesInDir {
			envStore = append(envStore, e)
		}

		// Now choose the confirmation
		if product == "confirm" { // if request is to confirm
			// Ask for confirmation
			confirm := methods.YesOrNoConfirmation()

			// What was the confirmation
			if confirm == "y" {  // yes
				log.Println("Continuing with the installtion of version: " + arguments.RequestedInstallVersion)
			} else { // no
				log.Println("Cancelling the installation...")
				os.Exit(0)
			}
		} else { // else choose

			// What is users choice
			choice := methods.Prompt_choice(len(envStore))

			// return the enviornment file to the main function
			choosenEnv := envStore[choice-1]
			return choosenEnv, nil

		}

	}

	return "", err
}

// Set Environment of the shell
func SetVersionEnv(filename string) error {

	log.Println("Attempting to open a terminal, after setting the environment of this installation.")

	// User Home
	usersHomeDir := os.Getenv("HOME")

	// Create a temp file to execute
	executeFile := arguments.TempDir + "openterminal.sh"
	_ = methods.DeleteFile(executeFile)
	_ = methods.CreateFile(executeFile)

	// The command
	var cmd []string
	cmdString := "gnome-terminal --working-directory=\"" + usersHomeDir + "\" --tab -e 'bash -c \"echo \\\"Sourcing Envionment file: "+ filename + "\\\"; source "+ filename +"; exec bash\"'"
	cmd = append(cmd, cmdString)

	// Write to the file
	_ = methods.WriteFile(executeFile, cmd)
	_, err := exec.Command("/bin/sh", executeFile).Output()
	if err != nil { return nil }

	// Cleanup the file file.
	_ = methods.DeleteFile(executeFile)

	return nil
}
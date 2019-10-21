// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
//     https://github.com/direct-state-transfer/dst-go/NOTICE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var tempGoPath, tempProjectRoot string

var (
	requiredGoVersionMajor = uint8(1)  //Go1.x Required
	minimumGoVersionMinor  = uint8(10) //Go1.10 Minimum
)

func main() {

	if len(os.Args) < 1 {
		fmt.Println("No commands to execute")
	}

	tempGoPath = os.Getenv("GOPATH")
	tempProjectRoot = os.Getenv("GOPATH") + os.Getenv("DST_GO_PATH")

	err := checkGoEnv(requiredGoVersionMajor, minimumGoVersionMinor)
	if err != nil {
		fmt.Printf("Error in go environment : %s\n", err)
		os.Exit(1)
	}

	err = installDependencies()
	if err != nil {
		fmt.Printf("Error installing dependencies : %s\n", err)
		os.Exit(1)
	}

	packagesList, err := generatePackagesList()
	if err != nil {
		fmt.Printf("Error generating package list : %s", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		fetchVendoredPackages()
		install()

	case "test":
		fetchVendoredPackages()
		triggerUnitTests(packagesList, os.Args[2:])

	case "runWalkthrough":
		fetchVendoredPackages()
		buildWalkthrough()
		runWalkthrough(os.Args[2:])

	case "lint":
		performLint()
	}
}

func installDependencies() (err error) {

	fmt.Println("\n----------------------------------------------------------")
	fmt.Println("                  Installing Dependencies                 ")
	fmt.Println("----------------------------------------------------------")

	fmt.Printf("\ngovendor...\t")
	cmd := exec.Command("go", "get", "-u", "-v", "github.com/kardianos/govendor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed. Below errors are reported\n\n%s", output)
	}
	fmt.Printf("Successful")

	fmt.Printf("\ngometalinter.v2...\t")
	cmd = exec.Command("go", "get", "-u", "-v", "gopkg.in/alecthomas/gometalinter.v2")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed. Below errors are reported\n\n%s", output)
	}
	fmt.Printf("Successful\n\n")
	return nil
}

func checkGoEnv(requiredGoVersionMajor, minimumGoVersionMinor uint8) (err error) {

	var goVersionMajor, goVersionMinor uint8

	goVersion := runtime.Version()

	if strings.Contains(goVersion, "devel") {
		return fmt.Errorf("development version of golang detected - %s", goVersion)
	}
	_, _ = fmt.Sscanf(runtime.Version(), "go%d.%d", &goVersionMajor, &goVersionMinor)

	if goVersionMajor != requiredGoVersionMajor {
		return fmt.Errorf("detected go version - %s. Major version %d required", goVersion, requiredGoVersionMajor)
	}
	if goVersionMinor < minimumGoVersionMinor {
		return fmt.Errorf("detected go version - %s. Minor version atleast %d required", goVersion, minimumGoVersionMinor)
	}
	if tempGoPath == "" {
		return fmt.Errorf("gopath not found")
	}

	return nil
}

func fetchVendoredPackages() {
	fmt.Printf("\n----------------------------------------------------------\n")
	fmt.Println("                Fetching vendored libraries                  ")
	fmt.Printf("---------------------------------------------------------- \n\n")
	var filesList []string
	var err error

	//Delete all the existing vendored files
	root := filepath.Join(tempProjectRoot, "vendor")
	if vendorFile, err := os.Stat(root); os.IsNotExist(err) || !vendorFile.IsDir() {
		fmt.Println("Vendor directory missing or it is a file")
		os.Exit(1)
	}
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Name() != "vendor.json" && info.Name() != "vendor" {
			filesList = append(filesList, path)
		}
		return nil
	})

	cmd := exec.Command("rm", "-rf")
	cmd.Args = append(cmd.Args, filesList...)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Delete existing vendored packages failed - ", err)
		os.Exit(1)
	}

	//Fetch vendored libraries using tool
	cmd = exec.Command(filepath.Join(tempGoPath, "/bin/govendor"), "sync", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Govendor sync failed - ", err)
		os.Exit(1)
	}

	//Print the status of vendored libraries
	cmd = exec.Command(filepath.Join(tempGoPath, "/bin/govendor"), "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Govendor list failed - ", err)
		os.Exit(1)
	}
}

func generatePackagesList() (packagesToTest []string, err error) {

	packagesToTestByteArr, err := pipe(
		exec.Command("go", "list", "./..."),
		exec.Command("grep", "-v", "-e", "build", "-e", "contract_store", "-e", "mocks", "-e", "walkthrough"),
	)

	if err != nil {
		return []string{}, err
	}
	packagesToTest = strings.Fields(packagesToTestByteArr.String())
	return packagesToTest, nil
}

//build and install dst-go
func install() {
	binaryName := "dst-go"
	fmt.Printf("\n----------------------------------------------------------\n")
	fmt.Printf("\n                   Installing %s                      \n", binaryName)
	fmt.Printf("---------------------------------------------------------- \n\n")
	cmd := exec.Command("go", "install", "-v", "./"+binaryName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Build failed - ", err)
		os.Exit(1)
	}
	fmt.Println("Installed successfully at", (filepath.Join(tempGoPath, "/bin/dst-go")))
}

//run unit tests for all packages
func triggerUnitTests(packagesList []string, userTestOpts []string) {

	//enable coverage reporting and disable test resulting caching
	defaultTestOpts := []string{"-cover", "-count=1"}

	fmt.Printf("\n----------------------------------------------------------\n")
	fmt.Println("                 Performing Unit Tests                    ")
	fmt.Printf("---------------------------------------------------------- \n\n")
	fmt.Printf("Running unit test with following options : \nEnabled by Default\t:%s \nProvided by user\t:%s\n", defaultTestOpts, userTestOpts)
	var errStr []string
	for _, packageToTest := range packagesList {

		fmt.Println("\nTesting package -", packageToTest)
		cmd := exec.Command("go", "test")
		cmd.Args = append(cmd.Args, defaultTestOpts...)
		cmd.Args = append(cmd.Args, userTestOpts...)
		cmd.Args = append(cmd.Args, packageToTest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			errStr = append(errStr, err.Error())
		}
	}
	if len(errStr) > 0 {
		fmt.Println("\nUnit test failed. Error Message -", errStr)
		os.Exit(1)
	}

	fmt.Println("\nUnit tests passed.")
}

//build walkthrough
func buildWalkthrough() {
	fmt.Printf("\n----------------------------------------------------------\n")
	fmt.Println("                   Building Walkthrough                   ")
	fmt.Printf("---------------------------------------------------------- \n\n")
	cmd := exec.Command("go", "build", "-o", "./walkthrough/walkthrough", "-v", "./walkthrough")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Walkthrough build failed - ", err)
		os.Exit(1)
	}
	fmt.Println("Successful")
}

//Walkthrough
func runWalkthrough(flags []string) {
	fmt.Printf("\n----------------------------------------------------------\n")
	fmt.Println("                  Executing Walkthrough                   ")
	fmt.Printf("---------------------------------------------------------- \n\n")
	var err error
	fmt.Printf("Running walkthrough with following options provided by user\t:%s\n", flags)

	cmd := exec.Command("./walkthrough")
	cmd.Args = append(cmd.Args, flags...)
	cmd.Dir = tempProjectRoot + "walkthrough"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Walkthrough Test Failed - ", err)
		os.Exit(1)
	}
}

//perform Lint
func performLint() {

	fmt.Println("----------------------------------------------------------")
	fmt.Println("                     Linting the code                     ")
	fmt.Println("----------------------------------------------------------")
	fmt.Printf("\nInstalling Linters...\t")
	cmd := exec.Command(filepath.Join(tempGoPath, "/bin/gometalinter.v2"), "--install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed. Below warnings and errors are reported\n\n%s\n", output)
		os.Exit(1)
	} else {
		fmt.Printf("Successful")
	}

	fmt.Printf("\nRunning Linters...\t")

	//Configuration for the linters is stored in a json file.
	configFile := filepath.Join(tempProjectRoot, "build/linterConfig.json")

	cmd = exec.Command(filepath.Join(tempGoPath, "/bin/gometalinter.v2"), "--config="+configFile, "./...")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed. Below warnings and errors are reported\n\n%s\n", output)
		os.Exit(1)
	} else {
		fmt.Printf("Successful")
	}
}

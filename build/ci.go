// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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
	"go/build"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var goPath, goBin, projectRootDir string

var (
	requiredGoVersionMajor = uint8(1)  //Go1.x Required
	minimumGoVersionMinor  = uint8(11) //Go1.11 Minimum
)

func main() {

	var err error
	var workingDir string

	_, programFilePath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Printf("Error in fetching caller information : %s\n", err)
		os.Exit(1)
	}

	projectRootDir = path.Dir(path.Dir(programFilePath))

	workingDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("Error in fetching current working directory : %s\n", err)
		os.Exit(1)
	}

	if projectRootDir != workingDir {
		fmt.Printf("Error - Program can only be invoked from project root directory\n")
		os.Exit(1)
	}

	if len(os.Args) < 1 {
		fmt.Println("No commands to execute")
	}

	//the variable from build package gives the proper gopath
	//from environment variables if set, else default value (~/go)
	goPath = build.Default.GOPATH
	goBin = filepath.Join(goPath, "bin")

	err = checkGoEnv(requiredGoVersionMajor, minimumGoVersionMinor)
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
		fmt.Printf("Error generating package list : %s\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetchDependencies":
		fetchVendoredPackages()

	case "install":
		fetchVendoredPackages()
		install()

	case "ciInstall":
		install()

	case "test":
		fetchVendoredPackages()
		triggerUnitTests(packagesList, os.Args[2:])

	case "ciTest":
		triggerUnitTests(packagesList, os.Args[2:])

	case "runWalkthrough":
		fetchVendoredPackages()
		buildWalkthrough()
		runWalkthrough(os.Args[2:])

	case "ciRunWalkthrough":
		buildWalkthrough()
		runWalkthrough(os.Args[2:])

	case "lint":
		// fetchVendoredPackages()
		performLint(os.Args[2:])
	}
}

func installDependencies() (err error) {

	fmt.Println("\n----------------------------------------------------------")
	fmt.Println("                  Installing Dependencies                 ")
	fmt.Println("----------------------------------------------------------")

	fmt.Printf("\ngolangci-lint v1.21.0...\t")

	outputByteArr, err := pipe(
		// #nosec G204. To suppress gosec linter warning. The below command is safe to use.
		exec.Command("curl", "-sfL", "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"),
		// #nosec G204. To suppress gosec linter warning. The below command is safe to use.
		exec.Command("sh", "-s", "--", "-b", goBin, "v1.21.0"),
	)
	if err != nil {
		return fmt.Errorf("failed. Below errors are reported\n\n%s", outputByteArr.String())
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

	return nil
}

func fetchVendoredPackages() {
	fmt.Printf("\n----------------------------------------------------------\n")
	fmt.Println("                Fetching vendored libraries                  ")
	fmt.Printf("---------------------------------------------------------- \n\n")
	var err error

	//Add missing and remove unnecessary dependencies.
	// #nosec G204. To suppress gosec linter warning. The below command is safe to use.
	cmd := exec.Command("go", "mod", "tidy", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("go mod tidy failed with error - ", err)
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
	fmt.Println("Installed successfully at", (filepath.Join(goPath, "/bin/dst-go")))
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
	cmd := exec.Command("go", "build", "-v")
	cmd.Dir = filepath.Join(projectRootDir, "walkthrough")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Walkthrough build failed - ", err)
		os.Exit(1)
	}
	fmt.Println("Build Successful")
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
	cmd.Dir = filepath.Join(projectRootDir, "walkthrough")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Walkthrough Test Failed - ", err)
		os.Exit(1)
	}
}

//perform Lint
func performLint(flags []string) {

	fmt.Println("----------------------------------------------------------")
	fmt.Println("                     Linting the code                     ")
	fmt.Println("----------------------------------------------------------")
	fmt.Printf("\nRunning Linters...\t")

	// Configuration is fetched from the file .golangci.yml in project root directory.
	// #nosec G204. To suppress gosec linter warning. The below command is safe to use.
	cmd := exec.Command(filepath.Join(goBin, "golangci-lint"), "run", "./...")
	cmd.Args = append(cmd.Args, flags...)
	cmd.Dir = projectRootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed. Below warnings and errors are reported\n\n%s\n", output)
		os.Exit(1)
	} else {
		fmt.Printf("\n%s\n", output)
		fmt.Printf("Successful. No errors detected\n")
	}
}

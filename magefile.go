//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// Build builds the entire project
func Build(ctx context.Context) error {
	mg.Deps(Clean)
	mg.Deps(BuildFrontend, BuildBackend)
	return nil
}

// BuildFrontend builds the frontend application
func BuildFrontend(ctx context.Context) error {
	fmt.Println("Building frontend...")
	
	// Change to frontend directory
	if err := os.Chdir("frontend"); err != nil {
		return fmt.Errorf("failed to change to frontend directory: %w", err)
	}
	defer os.Chdir("..")
	
	// Install dependencies
	if err := sh.Run("npm", "install"); err != nil {
		return fmt.Errorf("failed to install frontend dependencies: %w", err)
	}
	
	// Build frontend
	if err := sh.Run("npm", "run", "build"); err != nil {
		return fmt.Errorf("failed to build frontend: %w", err)
	}
	
	// Copy dist to root directory
	if err := sh.Run("cp", "-r", "dist", "../dist"); err != nil {
		return fmt.Errorf("failed to copy dist directory: %w", err)
	}
	
	fmt.Println("Frontend build completed successfully!")
	return nil
}

// BuildBackend builds the backend application
func BuildBackend(ctx context.Context) error {
	fmt.Println("Building backend...")
	
	// Change to backend directory
	if err := os.Chdir("backend"); err != nil {
		return fmt.Errorf("failed to change to backend directory: %w", err)
	}
	defer os.Chdir("..")
	
	// Download dependencies
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy go modules: %w", err)
	}
	
	// Build backend
	env := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
	}
	if err := sh.RunWith(env, "go", "build", "-a", "-installsuffix", "cgo", "-o", "../iptables-backend", "."); err != nil {
		return fmt.Errorf("failed to build backend: %w", err)
	}
	
	fmt.Println("Backend build completed successfully!")
	return nil
}

// Dev starts the development environment
func Dev(ctx context.Context) error {
	fmt.Println("Starting development environment...")
	
	// Start backend in background
	go func() {
		fmt.Println("Starting backend development server...")
		os.Chdir("backend")
		sh.Run("go", "run", "main.go")
	}()
	
	// Start frontend
	fmt.Println("Starting frontend development server...")
	os.Chdir("frontend")
	return sh.Run("npm", "run", "dev")
}

// DevFrontend starts only the frontend development server
func DevFrontend(ctx context.Context) error {
	fmt.Println("Starting frontend development server...")
	
	if err := os.Chdir("frontend"); err != nil {
		return fmt.Errorf("failed to change to frontend directory: %w", err)
	}
	
	// Install dependencies if needed
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		if err := sh.Run("npm", "install"); err != nil {
			return fmt.Errorf("failed to install frontend dependencies: %w", err)
		}
	}
	
	return sh.Run("npm", "run", "dev")
}

// DevBackend starts only the backend development server
func DevBackend(ctx context.Context) error {
	fmt.Println("Starting backend development server...")
	
	if err := os.Chdir("backend"); err != nil {
		return fmt.Errorf("failed to change to backend directory: %w", err)
	}
	
	// Tidy modules
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy go modules: %w", err)
	}
	
	return sh.Run("go", "run", "main.go")
}

// Docker builds and runs the Docker container
func Docker(ctx context.Context) error {
	mg.Deps(Build)
	
	fmt.Println("Building Docker image...")
	if err := sh.Run("docker", "build", "-t", "iptables-management", "."); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}
	
	fmt.Println("Running Docker container...")
	return sh.Run("docker", "run", "-p", "8080:8080", "--name", "iptables-management-container", "iptables-management")
}

// DockerBuild only builds the Docker image
func DockerBuild(ctx context.Context) error {
	mg.Deps(Build)
	
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", "iptables-management", ".")
}

// Clean removes build artifacts
func Clean(ctx context.Context) error {
	fmt.Println("Cleaning build artifacts...")
	
	// Remove backend binary
	if err := sh.Rm("iptables-backend"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove backend binary: %w", err)
	}
	
	// Remove dist directory
	if err := sh.Rm("dist"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove dist directory: %w", err)
	}
	
	// Clean frontend build
	frontendDist := filepath.Join("frontend", "dist")
	if err := sh.Rm(frontendDist); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove frontend dist: %w", err)
	}
	
	fmt.Println("Clean completed!")
	return nil
}

// Install installs all dependencies
func Install(ctx context.Context) error {
	fmt.Println("Installing dependencies...")
	
	// Install frontend dependencies
	fmt.Println("Installing frontend dependencies...")
	if err := os.Chdir("frontend"); err != nil {
		return fmt.Errorf("failed to change to frontend directory: %w", err)
	}
	if err := sh.Run("npm", "install"); err != nil {
		return fmt.Errorf("failed to install frontend dependencies: %w", err)
	}
	os.Chdir("..")
	
	// Install backend dependencies
	fmt.Println("Installing backend dependencies...")
	if err := os.Chdir("backend"); err != nil {
		return fmt.Errorf("failed to change to backend directory: %w", err)
	}
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to install backend dependencies: %w", err)
	}
	os.Chdir("..")
	
	fmt.Println("Dependencies installed successfully!")
	return nil
}

// Test runs all tests
func Test(ctx context.Context) error {
	fmt.Println("Running tests...")
	
	// Run backend tests
	fmt.Println("Running backend tests...")
	if err := os.Chdir("backend"); err != nil {
		return fmt.Errorf("failed to change to backend directory: %w", err)
	}
	if err := sh.Run("go", "test", "./..."); err != nil {
		fmt.Println("Backend tests failed or no tests found")
	}
	os.Chdir("..")
	
	// Run frontend tests
	fmt.Println("Running frontend tests...")
	if err := os.Chdir("frontend"); err != nil {
		return fmt.Errorf("failed to change to frontend directory: %w", err)
	}
	if err := sh.Run("npm", "test"); err != nil {
		fmt.Println("Frontend tests failed or no tests found")
	}
	os.Chdir("..")
	
	fmt.Println("Tests completed!")
	return nil
}

// Lint runs linting for both frontend and backend
func Lint(ctx context.Context) error {
	fmt.Println("Running linting...")
	
	// Lint backend
	fmt.Println("Linting backend...")
	if err := os.Chdir("backend"); err != nil {
		return fmt.Errorf("failed to change to backend directory: %w", err)
	}
	if err := sh.Run("go", "fmt", "./..."); err != nil {
		fmt.Println("Backend formatting failed")
	}
	if err := sh.Run("go", "vet", "./..."); err != nil {
		fmt.Println("Backend vetting failed")
	}
	os.Chdir("..")
	
	// Lint frontend
	fmt.Println("Linting frontend...")
	if err := os.Chdir("frontend"); err != nil {
		return fmt.Errorf("failed to change to frontend directory: %w", err)
	}
	if err := sh.Run("npm", "run", "lint"); err != nil {
		fmt.Println("Frontend linting failed or no lint script found")
	}
	os.Chdir("..")
	
	fmt.Println("Linting completed!")
	return nil
}
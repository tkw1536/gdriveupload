package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Options for a run of gdriveupload
type Options struct {
	CredsPath  string // Path to service account credentials
	FilePath   string // Path to file that should be uploaded
	FolderID   string // ID of folder to upload content in
	AllDrives  bool   // are we supporting all drives?
	OwnerEmail string // Email of user to set ownership for
}

func readOptions() (opts Options) {

	// read the arguments from command line
	flag.StringVar(&opts.CredsPath, "credspath", "", "Path to credentials")
	flag.StringVar(&opts.FilePath, "filepath", "", "Path of file to be uploaded")
	flag.StringVar(&opts.FolderID, "folderid", "", "ID of folder to upload file to")
	flag.StringVar(&opts.OwnerEmail, "owner", "", "Owner to transfer file to")
	flag.BoolVar(&opts.AllDrives, "alldrives", false, "Support all drives")
	flag.Parse()

	// check that there aren't any extra arguments
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Too many arguments.")
		flag.Usage()
		os.Exit(2)
	}

	// check that credspath exists
	if _, err := os.Stat(opts.CredsPath); err != nil || opts.CredsPath == "" {
		fmt.Fprintf(os.Stderr, "credspath %q does not exist.\n", opts.CredsPath)
		flag.Usage()
		os.Exit(2)
	}

	// check that filepath exists
	if _, err := os.Stat(opts.FilePath); err != nil || opts.FilePath == "" {
		fmt.Fprintf(os.Stderr, "filepath %q does not exist.\n", opts.FilePath)
		flag.Usage()
		os.Exit(2)
	}

	if opts.FolderID == "" {
		fmt.Fprintln(os.Stderr, "FolderID must not be empty.")
		flag.Usage()
		os.Exit(2)
	}

	return
}

func readCredentials(opts Options) (service *drive.Service, err error) {
	bytes, err := os.ReadFile(opts.CredsPath)
	if err != nil {
		return
	}

	var config *jwt.Config
	if config, err = google.JWTConfigFromJSON(bytes, drive.DriveScope); err != nil {
		return
	}
	config.Subject = opts.OwnerEmail

	client := config.Client(context.Background())
	return drive.NewService(context.Background(), option.WithHTTPClient(client))
}

func uploadFile(opts Options, service *drive.Service) (err error) {
	content, err := os.Open(opts.FilePath)
	if err != nil {
		return err
	}

	// prepare the filename
	_, name := filepath.Split(opts.FilePath)
	f := &drive.File{
		Name:    name,
		Parents: []string{opts.FolderID},
	}

	// create the file
	_, err = service.Files.Create(f).SupportsAllDrives(opts.AllDrives).Media(content).Do()
	return err
}

func main() {
	// read in the options from command line
	opts := readOptions()

	// generate credentials
	creds, err := readCredentials(opts)
	if err != nil {
		panic(err)
	}

	err = uploadFile(opts, creds)
	if err != nil {
		panic(err)
	}
}

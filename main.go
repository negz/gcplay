package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"gopkg.in/alecthomas/kingpin.v2"
)

func CreateNetwork(ctx context.Context, s *compute.NetworksService, project, name string) error {
	network := &compute.Network{Name: name}
	_, err := s.Insert(project, network).Context(ctx).Do()
	return err
}

func main() {
	var (
		app     = kingpin.New(filepath.Base(os.Args[0]), "A test harness.").DefaultEnvars()
		creds   = app.Flag("creds", "GCP credentials JSON.").Default("creds.json").ExistingFile()
		project = app.Arg("project", "GCP project.").String()
		name    = app.Arg("name", "GCP network name.").String()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	c, err := ioutil.ReadFile(*creds)
	kingpin.FatalIfError(err, "cannot read credentials file")

	ctx := context.Background()

	s, err := compute.NewService(ctx,
		option.WithCredentialsJSON(c),
		option.WithScopes(compute.ComputeScope))
	kingpin.FatalIfError(err, "cannot create compute networks service")

	kingpin.FatalIfError(CreateNetwork(ctx, s.Networks, *project, *name), "cannot create network")
}

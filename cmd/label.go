package cmd

import (
	"github.com/docker/engine-api/types"
	"github.com/nhurel/dim/lib"
	"github.com/spf13/cobra"
)

var labelCommand = &cobra.Command{
	Use:   "label [--delete] IMAGE[:TAG] LABEL_KEY=LABEL_VALUE...",
	Short: "Add / Remove a label to a given image",
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]
		labels := args[1:]

		var imageTags []string
		var tag string
		var err error

		var authConfig *types.AuthConfig
		if remoteFlag {
			if username != "" || password != "" {
				authConfig = &types.AuthConfig{Username: username, Password: password}
			}
			// TODO : get credentials the docker way and/or handle login
		}

		if pullFlag {
			if err = Dim.Pull(image); err != nil {
				return err
			}
		}

		if _, imageTags, err = Dim.GetImageInfo(image); err != nil {
			return err
		}

		if tag, err = guessTag(imageFlag, image, imageTags, overrideFlag); err != nil {
			return err
		}

		if deleteFlag {
			if err = Dim.RemoveLabel(image, labels, tag); err != nil {
				return err
			}
		} else {
			if err = Dim.AddLabel(image, labels, tag); err != nil {
				return err
			}
		}

		if overrideFlag {
			if dim.ListContains(imageTags, image) && image != tag {
				if err = Dim.Remove(image); err != nil {
					return err
				}
			}
		}

		if remoteFlag {
			if err = Dim.Push(tag, authConfig); err != nil {
				return err
			}
		}

		return nil
	},
}

var (
	imageFlag    string
	remoteFlag   bool
	overrideFlag bool
	pullFlag     bool
	deleteFlag   bool
)

func init() {
	labelCommand.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete the label")
	labelCommand.Flags().StringVarP(&imageFlag, "tag", "t", "", "Tag the new labeled image")
	labelCommand.Flags().BoolVarP(&remoteFlag, "remote", "r", false, "Delete the original image both locally and on the remote registry")
	labelCommand.Flags().BoolVarP(&overrideFlag, "override", "o", false, "Delete the original image locally only")
	labelCommand.Flags().BoolVarP(&pullFlag, "pull", "p", false, "Pull the image before adding label to ensure label is added to latest version")
	RootCommand.AddCommand(labelCommand)
}

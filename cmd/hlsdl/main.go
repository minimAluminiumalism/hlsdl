package main

import (
	"log"
	"os"
	"fmt"
	"github.com/canhlinh/hlsdl"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:          "hlsdl",
	RunE:         cmdF,
	SilenceUsage: true,
}

func main() {
	cmd.Flags().StringP("url", "u", "", "The manifest (m3u8) url")
	cmd.Flags().StringP("dir", "d", "./download", "The directory where the file will be stored")
	cmd.Flags().BoolP("record", "r", false, "Indicate whether the m3u8 is a live stream video and you want to record it")
	cmd.Flags().IntP("workers", "w", 2, "Number of workers to execute concurrent operations")
	cmd.SetArgs(os.Args[1:])

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func cmdF(command *cobra.Command, args []string) error {
	m3u8URL, err := command.Flags().GetString("url")
	if err != nil {
		fmt.Println("URL Error!")
		return err
	}

	dir, err := command.Flags().GetString("dir")
	if err != nil {
		return err
	}

	workers, err := command.Flags().GetInt("workers")
	if err != nil {
		return err
	}

	if record, err := command.Flags().GetBool("record"); err != nil {
		return err
	} else if record {
		return recordLiveStream(m3u8URL, dir)
	}

	return downloadVodMovie(m3u8URL, dir, workers)
}

func downloadVodMovie(url string, dir string, workers int) error {
	// Default value of `workers` is 2
	// media type:
	// url = "https://bitdash-a.akamaihd.net/content/sintel/hls/video/1500kbit.m3u8"
	// master type:
	// url = "http://devimages.apple.com.edgekey.net/iphone/samples/bipbop/bipbopall.m3u8"
	hlsDL := hlsdl.New(url, nil, dir, workers, true)
	filepath, err := hlsDL.Download()
	if err != nil {
		return err
	}
	log.Println("Downloaded file to " + filepath)
	return nil
}

func recordLiveStream(url string, dir string) error {
	recorder := hlsdl.NewRecorder(url, dir)
	recordedFile, err := recorder.Start()
	if err != nil {
		os.RemoveAll(recordedFile)
		return err
	}

	log.Println("Recorded file at ", recordedFile)
	return nil
}

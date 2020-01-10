package main

import (
	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"flag"
	"fmt"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"io"
	"os"
)

const (
	storageFilename = "test.flac"
	storageBucket   = "gcp-go-translation"
)

var (
	audioPath string
)

func init() {
	flag.StringVar(&audioPath, "p", "", "-p specify audio file path")
	flag.Parse()

	if audioPath == "" {
		fmt.Println("Please specify audio file path with -p")
		os.Exit(2)
	}
}

func main() {
	ctx := context.Background()

	// upload to gcp storage
	f, err := os.Open(audioPath)
	if nil != err {
		fmt.Println("Open audio file with error: ", err)
		os.Exit(2)
	}
	defer f.Close()

	// TODO: check before start to upload
	sc, err := storage.NewClient(ctx)
	if nil != err {
		fmt.Println("New storage client with error: ", err)
		os.Exit(2)
	}

	wc := sc.Bucket(storageBucket).Object(storageFilename).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		fmt.Println("Upload audio file with error: ", err)
		os.Exit(2)
	}
	if err := wc.Close(); err != nil {
		fmt.Println("Close storage client with error: ", err)
		os.Exit(2)
	}

	// create client
	client, err := speech.NewClient(ctx)
	if nil != err {
		fmt.Printf("Create client with error: %s\n", err)
		os.Exit(2)
	}

	op, err := client.LongRunningRecognize(ctx, &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:                            speechpb.RecognitionConfig_FLAC,
			SampleRateHertz:                     44100,
			LanguageCode:                        "en-US",
			AudioChannelCount:                   2,
			EnableSeparateRecognitionPerChannel: true,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: fmt.Sprintf("gs://%s/%s", storageBucket, storageFilename)},
		},
	})
	if nil != err {
		fmt.Printf("Recognition audio with error: %s\n", err)
		os.Exit(2)
	}

	resp, err := op.Wait(ctx)
	if nil != err {
		fmt.Println("Wait long running recognize with error: ", err)
		os.Exit(2)
	}

	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			fmt.Printf("%v\n", alt.Transcript)
		}
	}
}

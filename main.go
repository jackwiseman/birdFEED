package main

import (
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

const MinimumArea = 3000

func main() {
	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tmotion-detect [camera ID]")
		return
	}

	// parse args
	// deviceID := os.Args[1]

	stream := mjpeg.NewStream()

	// start capturing
	go mjpegCapture(stream)

	// start http server
	mux := http.NewServeMux()

	mux.Handle("/stream", stream)
	// mux.HandleFunc("/pics", picsHandler)

	fileServer := http.FileServer(http.Dir("./pics"))
	mux.Handle("/pics/", http.StripPrefix("/pics/", fileServer))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %s", err)
	}
}

func mjpegCapture(stream *mjpeg.Stream) {
	deviceID := 1

	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// window := gocv.NewWindow("Motion Window")
	// defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	status := "Ready"

	notified := false

	fmt.Printf("Start reading device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// rotate the image (webcam sits better upside down)
		gocv.Rotate(img, &img, gocv.Rotate180Clockwise)

		status = "Ready"
		statusColor := color.RGBA{0, 255, 0, 0}

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, &imgDelta)

		// remaining cleanup of the image to use for finding contours.
		// first use threshold
		gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

		// then dilate
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.Dilate(imgThresh, &imgThresh, kernel)
		kernel.Close()

		// now find contours
		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

		for i := 0; i < contours.Size(); i++ {
			area := gocv.ContourArea(contours.At(i))
			if area < MinimumArea {
				continue
			}

			// only want to trigger a motion detection event the FIRST time
			if !notified {
				gocv.IMWrite(fmt.Sprintf("./pics/%s.png", uniuri.New()), img)
				fmt.Println("bird spotted")
				notified = true
			}

			status = "Bird detected!"
			statusColor = color.RGBA{255, 0, 0, 0}
			gocv.DrawContours(&img, contours, i, statusColor, 2)

			rect := gocv.BoundingRect(contours.At(i))
			gocv.Rectangle(&img, rect, color.RGBA{0, 0, 255, 0}, 2)
		}

		contours.Close()

		if status == "Ready" {
			notified = false
		}

		gocv.PutText(&img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

		// needed for stream
		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf.GetBytes())
		buf.Close()
	}
}

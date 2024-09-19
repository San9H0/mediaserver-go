package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/v1/hls/:streamkey/output.m3u8", func(c echo.Context) error {
		fmt.Println("output.m3u8 streamkey:", c.Param("streamkey"))
		return c.File("output.m3u8")
	})
	e.GET("/v1/hls/:streamkey/:output", func(c echo.Context) error {
		fmt.Println("output streamkey :", c.Param("streamkey"), ", output:", c.Param("output"))
		return c.File(c.Param("output"))
	})
	e.Logger.Fatal(e.Start(":8080"))
}

func myhls() {
	//e := echo.New()
	//
	//playlist := m3u8.NewMasterPlaylist()
	////playlist.SetVersion(3)
	//
	//mediaPlayList, err := m3u8.NewMediaPlaylist(11, 11)
	//if err != nil {
	//	panic(err)
	//}
	//mediaPlayList.MediaType = m3u8.VOD
	//
	//playlist.Append("video_0.m3u8", mediaPlayList, m3u8.VariantParams{
	//	Bandwidth:  500,
	//	Resolution: "1280x720",
	//	FrameRate:  29.970,
	//	Codecs:     "avc1.42C020,Opus",
	//	//Audio:      "audio_0",
	//	//Alternatives: []*m3u8.Alternative{
	//	//	{
	//	//		Name:       "und",
	//	//		Language:   "und",
	//	//		Default:    true,
	//	//		Autoselect: "YES",
	//	//		Type:       "AUDIO",
	//	//		GroupId:    "audio_0",
	//	//		URI:        "audio-only.m3u8",
	//	//	},
	//	//},
	//})
	//
	//var mu sync.RWMutex
	//mediaPlayList.Slide("output_0.mp4", 2.542, "")
	//mediaPlayList.Slide("output_1.mp4", 3.085, "")
	//mediaPlayList.Slide("output_2.mp4", 2.577, "")
	//mediaPlayList.Slide("output_3.mp4", 3.142, "")
	//mediaPlayList.Slide("output_4.mp4", 3.058, "")
	//mediaPlayList.Slide("output_5.mp4", 3.123, "")
	//mediaPlayList.Slide("output_6.mp4", 3.152, "")
	//mediaPlayList.Slide("output_7.mp4", 3.088, "")
	//mediaPlayList.Slide("output_8.mp4", 2.271, "")
	//mediaPlayList.Slide("output_9.mp4", 2.194, "")
	//mediaPlayList.Slide("output_10.mp4", 3.192, "")
	//
	//mediaPlayList.Close()
	//
	//fmt.Println("playlist:", playlist.String())
	//
	//fmt.Println("mediaPlayList:", mediaPlayList.String())
	//
	//e.GET("/v1/hls/:streamkey/index.m3u8", func(c echo.Context) error {
	//	fmt.Println("index streamkey:", c.Param("streamkey"))
	//
	//	return c.String(200, playlist.String())
	//})
	//
	//e.GET("/v1/hls/:streamkey/video_0.m3u8", func(c echo.Context) error {
	//	fmt.Println("video_0 streamkey:", c.Param("streamkey"))
	//
	//	mu.RLock()
	//	str := mediaPlayList.String()
	//	mu.RUnlock()
	//	fmt.Println("[TESTDEBUG] return str:", str)
	//
	//	return c.String(200, str)
	//})
	//
	//e.GET("/v1/hls/:streamkey/:output", func(c echo.Context) error {
	//	fmt.Println("output streamkey :", c.Param("streamkey"), ", output:", c.Param("output"))
	//	return c.File(c.Param("output"))
	//})
	//
	//e.Logger.Fatal(e.Start(":8080"))
}

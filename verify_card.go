package admin

import (
	"Oaks/pkg/handlers"
	"Oaks/pkg/lib"
	"image"
	_ "image/png"
	"log"
	"net/http"
	"os"
)

func loadImageFromFile(imgPath string) (image.Image, error) {
	imageFile, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()

	img, _, err := image.Decode(imageFile)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func VerifyCard(w http.ResponseWriter, r *http.Request) {

	FilesUpload(w, r)

	/*
		client := gosseract.NewClient()
		defer client.Close()
		client.SetImage(handlers.Sess.Img_path)
		text, err := client.Text()
		if err != nil {
			log.Println("Can't get text from image!", err)
		}

		fmt.Println(text)


	*/

	log.Println("File path: ", handlers.Sess.Img_path)

}

func DisplayVerify(w http.ResponseWriter, r *http.Request) {

	data := lib.PageData{
		HeaderData: lib.Header{
			Title:   "OACS | Verify Personality",
			IsAdmin: handlers.Sess.IsAdmin,
		},
		FooterData: lib.Footer{
			CopyrightYear: 2023,
		},
	}

	lib.RenderPage(w, "card_verify.html", data)

}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/graphics-go/graphics"
	simplejson "github.com/bitly/go-simplejson"
)

type Content struct {
	Name         string    `json:"Name"`
	Texture      string    `json:"texture"`
	Rect         []int     `json:"rect"`
	Offset       []float64 `json:"offset"`
	OriginalSize []float64 `json:"originalSize"`
	CapInsets    []float64 `json:"capInsets"`
	Rotated      int       `json:"rotated"`
}

type Item struct {
	Type        string  `json:"__type__"`
	ContentData Content `json:"content" `
}

func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func TrimStringSpace(str string) string {
	strs := []string{"\n", " ", "\t", "\f", "\r", "\v"}
	for _, strItem := range strs {
		str = strings.Replace(str, strItem, "", -1)
	}

	return str
}

func getFilelist(path string) []string {
	var files []string
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		strLen := len(path)
		if strLen > 5 && path[strLen-5:] == ".json" {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	return files
}

func writeFile(fileName string, data []byte) {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		_, err = f.Write([]byte(data))
	}
}

// 保存Png图片
func saveImage(path string, img image.Image) (err error) {
	// 需要保存的文件
	imgfile, err := os.Create(path)
	defer imgfile.Close()

	// 以PNG格式保存文件
	err = png.Encode(imgfile, img)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func getImage() {

	//end := strings.Trim(, unicode.IsSpace)

	StrMap := make(map[string][]Content)

	files := getFilelist("./program")
	var jsonStrings []string
	for _, fileName := range files {
		strbytes, _ := ReadAll(fileName)
		str := string(strbytes)

		//处理空格Tab,回车
		str = TrimStringSpace(str)

		for {
			idx := strings.Index(str, "{\"__type__\":\"cc.SpriteFrame\"")
			if idx == -1 {
				break
			}

			str = str[idx:]
			index := strings.Index(str, "}}")
			if idx != -1 {
				jsonStrings = append(jsonStrings, str[:index+2])
			}

			str = str[index+2:]
		}
	}

	for _, strItem := range jsonStrings {

		var p1 Item
		err := json.Unmarshal([]byte(strItem), &p1) // 貌似这种解析方法需要提前知道 json 结构
		if err != nil {
			fmt.Println("err: ", err)
		}

		log.Println(p1.ContentData.Name, p1.ContentData.Texture)

		ctd := p1.ContentData
		if ctd.Rotated == 1 {
			ex := ctd.Rect[2]
			ctd.Rect[2] = ctd.Rect[3]
			ctd.Rect[3] = ex
		}

		if ary, ok := StrMap[ctd.Texture]; ok {
			StrMap[ctd.Texture] = append(ary, ctd)
		} else {
			ary = []Content{ctd}
			StrMap[ctd.Texture] = ary
		}

	}

	assets := getFilelist("./raw-assets")
	var picFiles []string
	for _, name := range assets {
		if len(name) > 4 && (name[len(name)-4:] == ".png") {
			picFiles = append(picFiles, name)
		}

		if len(name) > 4 && (name[len(name)-4:] == ".jpg") {
			log.Println(name)
		}
	}

	FileMap := make(map[string]string)
	for keyStr, _ := range StrMap {
		for _, fileName := range picFiles {

			if keyStr[:2] == fileName[11:13] {
				FileMap[keyStr] = fileName
			}
		}
	}

	ImageMap := make(map[string]image.Image)

	for keyStr, imgItem := range FileMap {
		game1, err := os.Open(imgItem)
		if err != nil {
			fmt.Println(err)
		}
		defer game1.Close()

		gameImg, _, err := image.Decode(game1) //解码
		if err != nil {
			fmt.Println(err)
		}

		ImageMap[keyStr] = gameImg
	}

	for keyStr, itemAry := range StrMap {
		switch ImageMap[keyStr].(type) {
		case *image.NRGBA:
			grgbImg := ImageMap[keyStr].(*image.NRGBA)
			for _, info := range itemAry {
				subImg := grgbImg.SubImage(image.Rect(info.Rect[0], info.Rect[1], info.Rect[0]+info.Rect[2], info.Rect[1]+info.Rect[3]))

				if info.Rotated == 1 {
					dst := image.NewRGBA(image.Rect(0, 0, info.Rect[3], info.Rect[2]))
					err := graphics.Rotate(dst, subImg, &graphics.RotateOptions{3 * math.Pi / 2})
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("./images/" + info.Name + ".png")
						saveImage("./images/"+info.Name+".png", dst)
					}

				} else {
					saveImage("./images/"+info.Name+".png", subImg)
				}

			}

		case *image.Paletted:
			grgbImg := ImageMap[keyStr].(*image.Paletted)
			for _, info := range itemAry {
				subImg := grgbImg.SubImage(image.Rect(info.Rect[0], info.Rect[1], info.Rect[0]+info.Rect[2], info.Rect[1]+info.Rect[3]))

				if info.Rotated == 1 {
					dst := image.NewRGBA(image.Rect(0, 0, info.Rect[3], info.Rect[2]))
					err := graphics.Rotate(dst, subImg, &graphics.RotateOptions{3 * math.Pi / 2})
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("./images/" + info.Name + ".png")
						saveImage("./images/"+info.Name+".png", dst)
					}

				} else {
					saveImage("./images/"+info.Name+".png", subImg)
				}
			}
		}
	}
}

func walk(node interface{}, depth int, outAry *[]map[string]interface{}, altasMap *map[string]string) {
	switch node.(type) {
	case []interface{}:
		ary := node.([]interface{})
		for _, item := range ary {
			walk(item, depth+1, outAry, altasMap)
		}
	case map[string]interface{}:
		mp := node.(map[string]interface{})
		(*outAry) = append((*outAry), mp)
		switch mp["__type__"] {
		case "cc.JsonAsset":
			data, _ := json.Marshal(mp["json"])
			var str bytes.Buffer
			_ = json.Indent(&str, []byte(data), "", "    ")
			writeFile("./out/"+mp["_name"].(string)+".json", str.Bytes())
		case "cc.Sprite":
			//fmt.Println(mp["__type__"], depth)
		case "cc.ScrollView":
			//fmt.Println(mp["__type__"], depth)
		case "cc.SpriteFrame":
			//fmt.Println(mp["__type__"], depth)
		case "cc.SpriteAtlas":
			altas := mp["_spriteFrames"].(map[string]interface{})
			for key, item := range altas {
				itm := item.(map[string]interface{})
				(*altasMap)[itm["__uuid__"].(string)] = key
			}
			//fmt.Println(mp["__type__"], depth)
		case "cc.AnimationClip":
			//fmt.Println(mp["__type__"], depth)
		case "cc.Node":
			//fmt.Println(mp["__type__"], depth)
		case "cc.Label":
			//fmt.Println(mp["__type__"], depth)
		case "cc.Animation":
			//fmt.Println(mp["__type__"], depth)
		case "cc.SceneAsset":
			//fmt.Println(mp["__type__"], mp["_name"], depth)
			//fmt.Println(mp["__type__"], mp["_name"], depth)
		case "cc.Scene":
			//fmt.Println(mp["__type__"], mp["_name"], depth)
			//fmt.Println(mp["__type__"], depth)
		case "cc.PrivateNode":
			//引擎內部
		case "cc.Prefab":
			//fmt.Println(mp["__type__"], mp["_name"], depth)
		case "cc.AudioClip":
			//fmt.Println(mp["__type__"], depth)
		case "cc.ProgressBar":
			//fmt.Println(mp["__type__"], depth)
		case "cc.RichText":
			//fmt.Println(mp["__type__"], depth)
		default:
			//fmt.Println("unhandle", mp["__type__"], depth)
		}

		if depth == 3 {
			if mp["__type__"] == nil {
				//fmt.Println(mp)
			} else {
				//fmt.Println(mp["__type__"], depth)
			}

		}

	}
}

func parse_file(data interface{}) {
	nodeAry := data.([]interface{})
	for _, dd := range nodeAry {
		switch dd.(type) {
		case []interface{}:
			dAry := dd.([]interface{})
			if reflect.TypeOf(dAry[0]).String() == "map[string]interface {}" {
				mp := dAry[0].(map[string]interface{})
				if mp["__type__"] == "cc.SceneAsset" {
					data, _ := json.Marshal(dAry)
					var str bytes.Buffer
					_ = json.Indent(&str, []byte(data), "", "    ")
					writeFile("./out/"+mp["_name"].(string)+".fire.json", str.Bytes())
				} else if mp["__type__"] == "cc.Prefab" {
					data, _ := json.Marshal(dAry)
					var str bytes.Buffer
					_ = json.Indent(&str, []byte(data), "", "    ")
					writeFile("./out/"+mp["_name"].(string)+".prefab.json", str.Bytes())
				}
			}
		case map[string]interface{}:
			// mp := dd.(map[string]interface{})
			// switch mp["__type__"] {
			// case "cc.SpriteAtlas":
			// 	fmt.Println(mp["__type__"])
			// }
		}
	}
}

func main() {
	spritefameMap := make(map[string]map[string]interface{})
	fileMap := make(map[string][]map[string]interface{})
	structMap := make(map[string]interface{})
	files := getFilelist("./program/res")

	uuidScriptMap := make(map[string]string)
	altasMap := make(map[string]string)

	strbytes, _ := ReadAll("./program/src/project.js")
	str := string(strbytes)

	//处理空格Tab,回车
	str = TrimStringSpace(str)
	var jsonStrings []string
	for {
		idx := strings.Index(str, "cc._RF.push(e")
		if idx == -1 {
			break
		}

		str = str[idx:]
		index := strings.Index(str, ")")
		if idx != -1 {
			jsonStrings = append(jsonStrings, str[:index+2])
			strs := strings.Split(str[:index+2], "\"")
			//fmt.Println(strs[1], strs[3])
			uuidScriptMap[strs[1]] = strs[3]
		}

		str = str[index+2:]
	}

	//fmt.Println(jsonStrings)

	for _, fileName := range files {
		fmt.Println(fileName)
		strbytes, _ := ReadAll(fileName)

		root, _ := simplejson.NewJson(strbytes)
		var outAry []map[string]interface{}
		walk(root.Interface(), 1, &outAry, &altasMap)
		structMap[fileName] = root.Interface()
		fileMap[fileName] = outAry
	}

	fmt.Println(altasMap)

	for _, item := range fileMap {
		for _, mp := range item {
			if mp["__type__"] == "cc.SpriteFrame" {
				if im, ok := mp["_name"]; ok {
					spritefameMap[im.(string)] = mp
				}

				if im, ok := mp["content"]; ok {
					cm := im.(map[string]interface{})
					spritefameMap[cm["name"].(string)] = mp
				}
			}
		}
	}

	for fileName, item := range fileMap {
		for _, mp := range item {
			if mp["__type__"] == "cc.SceneAsset" {
				parse_file(structMap[fileName])
			}
		}
	}
}

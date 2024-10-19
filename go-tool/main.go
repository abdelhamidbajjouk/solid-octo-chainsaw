package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitforge-srl/ptResource/go-tool/runtimedeps"
	"github.com/bitforge-srl/ptResource/go-tool/types"
)

var basePath string

func main() {
	var err error
	basePath, err = os.Executable()
	basePath = filepath.Dir(basePath)
	if err != nil {
		panic(err)
	}
	items, err := os.ReadDir(basePath)
	if err != nil {
		panic(err)
	}
	jsons := make([]string, 0)

	for _, item := range items {
		if filepath.Ext(item.Name()) == ".json" {
			jsons = append(jsons, item.Name())
		}
	}

	numJobs := len(jsons)
	jobs := make(chan string, numJobs)
	results := make(chan bool, numJobs)
	for i := range jsons {
		go makeTaraWorker(i, jobs, results)
	}

	for _, name := range jsons {
		jobs <- name
	}
	close(jobs)

	for a := 0; a < numJobs; a++ {
		<-results
	}

}

func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

// func makeTara(filename string) error {
// 	resourcesFolder := filepath.Dir(basePath)
// 	filenameFullPath := filepath.Join(basePath, filename)
// 	outResourceFolderFullPath := fileNameWithoutExtSliceNotation(filenameFullPath)
// 	if err := os.Mkdir(outResourceFolderFullPath, 0777); err != nil {
// 		//return err
// 	}

// 	jsonStruct := &types.JSON{}

// 	jsonFile, err := os.ReadFile(filenameFullPath)
// 	if err != nil {
// 		return err
// 	}

// 	if err := json.Unmarshal(jsonFile, jsonStruct); err != nil {
// 		return err
// 	}

// 	for _, resource := range jsonStruct.Items {
// 		if jsonStruct.ID == "PROPS" {
// 			propsSrcFolder := filepath.Join(resourcesFolder, resource.Src)
// 			entries, err := os.ReadDir(propsSrcFolder)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			for _, entry := range entries {
// 				if entry.IsDir() {
// 					runtimedeps.PackTaraExternal(filepath.Join(propsSrcFolder, entry.Name()), filepath.Join(outResourceFolderFullPath, entry.Name()))
// 				}
// 			}

// 			break
// 		}
// 		filenameSpl := strings.Split(resource.Src, "?")

// 		resFile, err := os.ReadFile(filepath.Join(resourcesFolder, filenameSpl[0]))
// 		if err != nil {
// 			log.Printf("Skipping file due to error: %s", err.Error())
// 			continue
// 		}

// 		outFile, err := os.Create(filepath.Join(outResourceFolderFullPath, resource.Name+filepath.Ext(filenameSpl[0])))
// 		if err != nil {
// 			panic(err)
// 		}
// 		outFile.Write(resFile)
// 		defer outFile.Close()
// 	}

// 	return runtimedeps.PackTaraExternal(outResourceFolderFullPath, filepath.Join(basePath, strings.ToLower(jsonStruct.ID)))

// }
func makeTara(filename string) error {
    resourcesFolder := filepath.Dir(basePath)
    filenameFullPath := filepath.Join(basePath, filename)
    outResourceFolderFullPath := fileNameWithoutExtSliceNotation(filenameFullPath)
    
    if err := os.Mkdir(outResourceFolderFullPath, 0777); err != nil {
        return err
    }

    jsonStruct := &types.JSON{}
    jsonFile, err := os.ReadFile(filenameFullPath)
    if err != nil {
        return err
    }
    
    if err := json.Unmarshal(jsonFile, jsonStruct); err != nil {
        return err
    }
    
    for _, resource := range jsonStruct.Items {
        switch jsonStruct.ID {
        case "PROPS", "MAPS":
            // Pack props/maps into TARA
            propsSrcFolder := filepath.Join(resourcesFolder, resource.Src)
            runtimedeps.PackTaraExternal(propsSrcFolder, outResourceFolderFullPath)

        case "IMAGES", "SOUNDS", "MODELS":
            // Store raw files in separate folders (no TARA)
            rawFolder := filepath.Join(resourcesFolder, jsonStruct.ID)
            if _, err := os.Stat(rawFolder); os.IsNotExist(err) {
                os.Mkdir(rawFolder, 0777)
            }
            filenameSpl := strings.Split(resource.Src, "?")
            resFile, err := os.ReadFile(filepath.Join(resourcesFolder, filenameSpl[0]))
            if err != nil {
                log.Printf("Skipping file due to error: %s", err.Error())
                continue
            }
            outFile, err := os.Create(filepath.Join(rawFolder, resource.Name+filepath.Ext(filenameSpl[0])))
            if err != nil {
                return err
            }
            outFile.Write(resFile)
            outFile.Close()
        }
    }

    if jsonStruct.ID == "PROPS" || jsonStruct.ID == "MAPS" {
        return runtimedeps.PackTaraExternal(outResourceFolderFullPath, filepath.Join(basePath, strings.ToLower(jsonStruct.ID)))
    }

    return nil
}

func makeTaraWorker(id int, jobs <-chan string, results chan<- bool) {
	for j := range jobs {
		fmt.Println("worker", id, "started  job", j)
		err := makeTara(j)
		if err != nil {
			panic(err)
		}
		fmt.Println("worker", id, "finished job", j)
		results <- true
	}
}

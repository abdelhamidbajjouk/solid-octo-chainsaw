package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"

    "crypto/md5"

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

func makeTara(filename string) error {
    resourcesFolder := filepath.Dir(basePath)
    filenameFullPath := filepath.Join(basePath, filename)
    outResourceFolderFullPath := fileNameWithoutExtSliceNotation(filenameFullPath)
    if err := os.Mkdir(outResourceFolderFullPath, 0777); err != nil && !os.IsExist(err) {
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
        resourcePath := filepath.Join(resourcesFolder, resource.Src)

        switch resource.Type {
        case ResourceTypeImage:
            // Generate hash for image
            hash, err := generateHash(resourcePath)
            if err != nil {
                log.Printf("Error generating hash for image %s: %v", resourcePath, err)
                continue
            }
            resource.Hash = hash

            // Create image directory if it doesn't exist
            imageDir := filepath.Join(outResourceFolderFullPath, "images")
            if err := os.MkdirAll(imageDir, 0777); err != nil && !os.IsExist(err) {
                return err
            }

            // Copy image file with resource ID (including hash) as name
            outImagePath := filepath.Join(imageDir, fmt.Sprintf("%s_%s%s", resource.ID, hash, filepath.Ext(resourcePath)))

            err = copyFile(resourcePath, outImagePath)
            if err != nil {
                log.Printf("Error copying image %s: %v", resourcePath, err)
                continue
            }

        case ResourceTypeSound:
            // Create sound directory if it doesn't exist
            soundDir := filepath.Join(outResourceFolderFullPath, "sounds")
            if err := os.MkdirAll(soundDir, 0777); err != nil && !os.IsExist(err) {
                return err
            }

            // Copy sound file with resource ID as name
            outSoundPath := filepath.Join(soundDir, resource.ID+filepath.Ext(resourcePath))
            err = copyFile(resourcePath, outSoundPath)
            if err != nil {
                log.Printf("Error copying sound %s: %v", resourcePath, err)
                continue
            }

        case ResourceTypeModel:
            // Create model directory if it doesn't exist
            modelDir := filepath.Join(outResourceFolderFullPath, "models")
            if err := os.MkdirAll(modelDir, 0777); err != nil && !os.IsExist(err) {
                return err
            }

            // Copy model file with resource ID as name
            outModelPath := filepath.Join(modelDir, resource.ID+filepath.Ext(resourcePath))
            err = copyFile(resourcePath, outModelPath)
            if err != nil {
                log.Printf("Error copying model %s: %v", resourcePath, err)
                continue
            }
        }
    }

    return runtimedeps.PackTaraExternal(outResourceFolderFullPath, filepath.Join(basePath, strings.ToLower(jsonStruct.ID)))
}

func generateHash(filePath string) (string, error) {
    // Read the file contents
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    // Create a MD5 hash
    h := md5.New()
    if _, err := h.Write(data); err != nil {
        return "", err
    }
    hash := fmt.Sprintf("%x", h.Sum(nil))

    return hash, nil
}

func copyFile(src, dst string) error {
    in, err := ioutil.ReadFile(src)
    if err != nil {
        return err
    }

    return ioutil.WriteFile(dst, in, 0644)
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
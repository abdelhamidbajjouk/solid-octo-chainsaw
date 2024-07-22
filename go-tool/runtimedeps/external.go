package runtimedeps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func PackTaraExternal(inFolder string, outPath string) error {

	exPath, err := os.Executable()
	if err != nil {
		return err
	}
	basePath := filepath.Dir(exPath)
	jarPath := filepath.Join(basePath, "runtimedeps", "TaraTool.jar")

	cmd := exec.Command("java", "-jar", jarPath, "pack", inFolder, fmt.Sprintf("%s.tara", outPath))
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

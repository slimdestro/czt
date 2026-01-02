package utils

import (
	"os"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	t.Run("File Not Found", func(t *testing.T) {
		LoadEnv("non_existent_file")
	})

	t.Run("Valid Env File", func(t *testing.T) {
		content := "KEY1=VALUE1\n# Comment line\n\nKEY2=VALUE2=WITH_EQUALS\nINVALIDLINE"
		tmpFile, err := os.CreateTemp("", ".env")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		LoadEnv(tmpFile.Name())

		if os.Getenv("KEY1") != "VALUE1" {
			t.Errorf("expected VALUE1, got %s", os.Getenv("KEY1"))
		}
		if os.Getenv("KEY2") != "VALUE2=WITH_EQUALS" {
			t.Errorf("expected VALUE2=WITH_EQUALS, got %s", os.Getenv("KEY2"))
		}
	})
}

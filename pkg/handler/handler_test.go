package handler

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResolveFolderWithoutAnnotation(t *testing.T) {
	expectedFolder := "/tmp"

	var o metav1.Object
	o = &metav1.ObjectMeta{
		Annotations: map[string]string{
			"differenct-annotation.io/folder": "/tmp/different",
		},
	}

	folder := ResolveFolder(HandlerConfig{
		FolderAnnotation: "config-collector.io/folder",
		Folder:           expectedFolder,
	}, o)

	assert.Equal(t, folder, expectedFolder)
}

func TestResolveFolderWithAnnotation(t *testing.T) {
	annotation := "config-collector.io/folder"
	expectedFolder := "/tmp/different"

	var o metav1.Object
	o = &metav1.ObjectMeta{
		Annotations: map[string]string{
			annotation: expectedFolder,
		},
	}

	folder := ResolveFolder(HandlerConfig{
		FolderAnnotation: annotation,
		Folder:           "/tmp",
	}, o)

	assert.Equal(t, folder, expectedFolder)
}

func TestResolveFolderRemoveTrailingSlash(t *testing.T) {
	expectedFolder := "/tmp"

	var o metav1.Object
	o = &metav1.ObjectMeta{
		Annotations: map[string]string{
			"differenct-annotation.io/folder": "/tmp/different",
		},
	}

	folder := ResolveFolder(HandlerConfig{
		FolderAnnotation: "config-collector.io/folder",
		Folder:           expectedFolder + "/",
	}, o)

	assert.Equal(t, folder, expectedFolder)
}

func TestResolveFilePath(t *testing.T) {
	expectedPath := "/tmp/test"

	var o metav1.Object
	o = &metav1.ObjectMeta{}

	path := ResolveFilePath(HandlerConfig{
		FolderAnnotation: "config-collector.io/folder",
		Folder:           "/tmp",
	}, o, "test")

	assert.Equal(t, path, expectedPath)
}

func TestNewFileResolverAdd(t *testing.T) {

	tmpDir, err := ioutil.TempDir("/tmp", "test")
	if err != nil {
		panic(err)
	}
	config := HandlerConfig{
		Folder: tmpDir,
	}

	handler := NewFileHandler(config)

	cm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cm",
		},
		Data: map[string]string{
			"test": ":)",
		},
	}

	filename := ResolveFilePath(config, cm.GetObjectMeta(), "test")

	// Create file and check it exists
	handler.AddFunc(cm)
	_, err = os.Stat(filename)
	assert.Assert(t, err == nil)
}

func TestNewFileResolverDelete(t *testing.T) {

	tmpDir, err := ioutil.TempDir("/tmp", "test")
	if err != nil {
		panic(err)
	}
	config := HandlerConfig{
		Folder: tmpDir,
	}

	handler := NewFileHandler(config)

	cm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cm",
		},
		Data: map[string]string{
			"test": ":)",
		},
	}

	filename := ResolveFilePath(config, cm.GetObjectMeta(), "test")
	ioutil.WriteFile(filename, []byte("test"), 0644)
	// Delete file and check is gone
	handler.DeleteFunc(cm)
	_, err = os.Stat(filename)
	assert.Assert(t, err != nil)
}

func TestNewFileResolver(t *testing.T) {

	tmpDir, err := ioutil.TempDir("/tmp", "test")
	if err != nil {
		panic(err)
	}
	config := HandlerConfig{
		Folder: tmpDir,
	}

	handler := NewFileHandler(config)

	cm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cm",
		},
		Data: map[string]string{
			"test": ":)",
		},
	}

	testFilename := ResolveFilePath(config, cm.GetObjectMeta(), "test")
	deleteThis := ResolveFilePath(config, cm.GetObjectMeta(), "delete.this")
	createThis := ResolveFilePath(config, cm.GetObjectMeta(), "create.this")
	ioutil.WriteFile(testFilename, []byte("test"), 0644)
	// Delete file and check is gone
	handler.UpdateFunc(&apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cm",
		},
		Data: map[string]string{
			"test":        ":)",
			"delete.this": "This should be deleted",
		},
	},
		&apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cm",
			},
			Data: map[string]string{
				"test":        ":(",
				"create.this": "This is a new file",
			},
		})
	_, err = os.Stat(testFilename)
	assert.Assert(t, err == nil)
	_, err = os.Stat(createThis)
	assert.Assert(t, err == nil)
	_, err = os.Stat(deleteThis)
	assert.Assert(t, err != nil)
}

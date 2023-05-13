package email

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-fs"
)

func TestAttachment_MarshalJSON(t *testing.T) {
	j, err := json.Marshal(Attachment{
		PartID:      "PartID",
		ContentID:   "ContentID",
		ContentType: "ContentType",
		MemFile: fs.MemFile{
			FileName: "FileName",
			FileData: []byte("FileData"),
		},
	})
	require.NoError(t, err, "json.Marshal")
	require.Equal(t, `{"partID":"PartID","contentID":"ContentID","contentType":"ContentType","filename":"FileName","data":"RmlsZURhdGE="}`, string(j))
}

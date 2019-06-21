// Package uuidsubdirs provides functions to split up a uu.ID
// into a series of sub-directories so that an unlimited number
// of UUIDs can be used as directories.
//
// Example:
//   The UUID f0498fad-437c-4954-ad82-8ec2cc202628 maps to the path
//   f04/98f/ad43/7c4954/ad828ec2cc202628
package uuidsubdirs

import (
	"fmt"
	"os"
	"path"
	"strings"

	command "github.com/ungerik/go-command"
	fs "github.com/ungerik/go-fs"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/uu"
)

// SplitUUID splits a uu.ID into 5 sub-strings.
// Example:
//   The UUID f0498fad-437c-4954-ad82-8ec2cc202628 maps to the path
//   f04/98f/ad43/7c4954/ad828ec2cc202628
func SplitUUID(id uu.ID) []string {
	// f0498fad-437c-4954-ad82-8ec2cc202628
	// f04/98f/ad43/7c4954/ad828ec2cc202628
	hex := id.Hex()
	return []string{
		hex[0:2],
		hex[2:5],
		hex[5:8],
		hex[8:16],
		hex[16:32],
	}
}

var PathForUUIDArgs struct {
	command.ArgsDef

	ID uu.ID `arg:"id"`
}

func PathForUUID(id uu.ID) string {
	return path.Join(SplitUUID(id)...)
}

func DirForUUID(baseDir fs.File, id uu.ID) (idDir fs.File) {
	return baseDir.Join(SplitUUID(id)...)
}

var PathToUUIDArgs struct {
	command.ArgsDef

	Path string `arg:"path"`
}

func PathToUUID(path string) (string, error) {
	path = strings.Trim(path, string(os.PathSeparator))
	parts := strings.Split(path, string(os.PathSeparator))
	n := len(parts)
	if n < 5 {
		return "", errors.Errorf("path does not have enough parts to build a UUID: '%s'", path)
	}
	idString := strings.Join(parts[n-5:], "")
	if len(idString) != 32 {
		return "", errors.Errorf("last 5 path parts must have a length of 32 characters, but has %d", len(idString))
	}
	id, err := uu.IDFromString(idString)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func UUIDFromDir(baseDir, idDir fs.File) (id uu.ID, err error) {
	if !idDir.Exists() {
		return uu.IDNil, fs.NewErrDoesNotExist(idDir)
	}
	if !idDir.IsDir() {
		return uu.IDNil, fs.NewErrIsNotDirectory(idDir)
	}
	idString := idDir.PathWithSlashes()
	idString = strings.TrimPrefix(idString, baseDir.PathWithSlashes())
	idString = strings.ReplaceAll(idString, "/", "")
	if len(idString) != 32 {
		return uu.IDNil, errors.Errorf("sub-path '%s' can't be converted to a UUID. Full path: '%s'", string(idString), string(idDir))
	}
	id, err = uu.IDFromString(idString)
	if err != nil {
		return uu.IDNil, errors.Errorf("sub-path is not a valid UUID: '%s'", string(idDir))
	}
	return id, nil
}

func EnumUUIDDirs(baseDir fs.File, callback func(idDir fs.File, id uu.ID) error) error {
	return baseDir.ListDir(func(level0Dir fs.File) error {
		if !level0Dir.Exists() || level0Dir.IsHidden() {
			return nil
		}
		if !level0Dir.IsDir() {
			fmt.Println("Directory expected but found file:", level0Dir)
			return nil
		}
		return level0Dir.ListDir(func(level1Dir fs.File) error {
			if !level1Dir.Exists() || level1Dir.IsHidden() {
				return nil
			}
			if !level1Dir.IsDir() {
				fmt.Println("Directory expected but found file:", level1Dir)
				return nil
			}
			return level1Dir.ListDir(func(level2Dir fs.File) error {
				if !level2Dir.Exists() || level2Dir.IsHidden() {
					return nil
				}
				if !level2Dir.IsDir() {
					fmt.Println("Directory expected but found file:", level2Dir)
					return nil
				}
				return level2Dir.ListDir(func(level3Dir fs.File) error {
					if !level3Dir.Exists() || level3Dir.IsHidden() {
						return nil
					}
					if !level3Dir.IsDir() {
						fmt.Println("Directory expected but found file:", level3Dir)
						return nil
					}
					return level3Dir.ListDir(func(idDir fs.File) error {
						if !idDir.Exists() || idDir.IsHidden() {
							return nil
						}
						if !idDir.IsDir() {
							fmt.Println("Directory expected but found file:", idDir)
							return nil
						}
						id, err := UUIDFromDir(baseDir, idDir)
						if err != nil {
							return err
						}
						return callback(idDir, id)
					})
				})
			})
		})
	})
}

func DeleteUpTo(baseDir, dir fs.File) error {
	basePath := baseDir.Path()
	if dir.Path() == basePath {
		return nil
	}
	// fmt.Println("deleting", dir.Path())
	err := dir.RemoveRecursive()
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		dir = dir.Dir()
		if dir.Path() == basePath || !dir.IsEmptyDir() {
			return nil
		}
		// fmt.Println("deleting", dir.Path())
		err = dir.Remove()
		if err != nil {
			return err
		}
	}
	return nil
}

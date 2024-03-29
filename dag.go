package merkledag

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sort"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func sliceFile(file []byte, chunkSize int) ([][]byte, error) {
	var chunks [][]byte
	reader := bytes.NewReader(file)
	chunk := make([]byte, chunkSize)

	for {
		n, err := reader.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if n > 0 {
			chunks = append(chunks, append([]byte(nil), chunk[:n]...))
		}
	}

	return chunks, nil
}

func orderDir(dir Dir) Dir {
	links := dir.Links()
	sort.Slice(links, func(i, j int) bool {
		return links[i].Name < links[j].Name
	})
	dir.SetLinks(links)

	return dir
}

func Add(store KVStore, node Node, kvstore KVStore) (string, error) {
	var hashes []string

	var traverse func(Node) (string, error)
	traverse = func(n Node) (string, error) {
		switch v := n.(type) {
		case File:
			chunks, err := sliceFile(v, 1024)
			if err != nil {
				return "", err
			}

			hasher := sha256.New()
			for _, chunk := range chunks {
				hasher.Write(chunk)
			}
			fileHash := hex.EncodeToString(hasher.Sum(nil))

			err = kvstore.Put(fileHash, v)
			if err != nil {
				return "", err
			}

			hashes = append(hashes, fileHash)
			return fileHash, nil

		case Dir:
			dir := orderDir(v)

			it := dir.It()
			for it.HasNext() {
				subNode, err := it.Next()
				if err != nil {
					return "", err
				}

				subHash, err := traverse(subNode)
				if err != nil {
					return "", err
				}
				hashes = append(hashes, subHash)
			}

			sort.Strings(hashes)

			hasher := sha256.New()
			for _, h := range hashes {
				hasher.Write([]byte(h))
			}
			dirHash := hex.EncodeToString(hasher.Sum(nil))
			hashes = hashes[:0]

			err = kvstore.Put(dirHash, v)
			if err != nil {
				return "", err
			}

			return dirHash, nil
		}

		return "", ErrUnknownNodeType
	}

	rootHash, err := traverse(node)
	if err != nil {
		return "", err
	}

	if len(hashes) == 0 {
		return rootHash, nil
	}

	hasher := sha256.New()
	hasher.Write([]byte(rootHash))
	for _, h := range hashes {
		hasher.Write([]byte(h))
	}
	merkleRoot := hex.EncodeToString(hasher.Sum(nil))

	return merkleRoot, nil
}

// ErrUnknownNodeType is the error returned when encountering an unknown node type.
var ErrUnknownNodeType = errors.New("unknown node type")

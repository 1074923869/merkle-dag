package merkledag

import "hash"

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}
type node struct{
	Type merkleNodetype;
	Data []byte;
	Children []*merkleChildNode;//子节点表
	Hash byte[];

}
func sliceFile(file File, chunkSize int) ([][]byte, error) {  
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
			chunks = append(chunks, chunk[:n])  
		}  
	}  
  
	return chunks, nil  
}  
func orderDir(dir Dir) Dir {  
	func orderDir(dir Dir) Dir {
		
		links := dir.Links()
		sort.Slice(links, func(i, j int) bool {
			return links[i].Name < links[j].Name
		})
		dir.SetLinks(links)
	
	return dir  
}  
func Add(store KVStore, node Node, h hash.Hash) []byte {
	// TODO 将分片写入到KVStore中，并返回Merkle Root
	var hashes []string  
  
	var traverse func(Node) (string, error)  
	traverse = func(n Node) (string, error) {  
		switch v := n.(type) {  
		case File:  
			// 对文件内容进行切片  
			chunks, err := sliceFile(v, 1024) // 假设每个块的大小为1024字节  
			if err != nil {  
				return "", err  
			}  
  
			// 计算每个块的哈希值，并拼接为文件的哈希值  
			hasher := sha256.New()  
			for _, chunk := range chunks {  
				hasher.Write(chunk)  
			}  
			fileHash := hex.EncodeToString(hasher.Sum(nil))  
  
			// 保存文件到KVStore  
			err = kvstore.Put(fileHash, v)  
			if err != nil {  
				return "", err  
			}  
  
			hashes = append(hashes, fileHash)  
			return fileHash, nil  
  
		case Dir:  
			// 对文件夹的子节点进行排序  
			dir := orderDir(v)  
  
			// 遍历文件夹的子节点  
			it := dir.It()  
			for it.HasNext() {  
				subNode, err := it.Next()  
				if err != nil {  
					return "", err  
				}  
  
				// 递归调用traverse函数  
				subHash, err := traverse(subNode)  
				if err != nil {  
					return "", err  
				}  
				hashes = append(hashes, subHash)  
			}  
  
			// 对子节点哈希值排序  
			sort.Strings(hashes)  
  
			// 计算文件夹的哈希值（将排序后的子节点哈希值连接起来并哈希）  
			hasher := sha256.New()  
			for _, h := range hashes {  
				hasher.Write([]byte(h))  
			}  
			dirHash := hex.EncodeToString(hasher.Sum(nil))  
			hashes = hashes[:0] // 重置hashes切片，为下一个文件夹做准备  

			
	// 保存文件夹到KVStore（这里假设KVStore有一个Put方法）  
			err = kvstore.Put(dirHash, v)  
			if err != nil {  
				return "", err  
			}  
  
			return dirHash, nil  
		}  
  
		return "", ErrUnknownNodeType  
	}  
  
	// 从根节点开始遍历  
	rootHash, err := traverse(node)  
	if err != nil {  
		return "", err  
	}  
  
	// 如果没有其他节点，则根节点的哈希值就是Merkle Root  
	if len(hashes) == 0 {  
		return rootHash, nil  
	}  
  
	// 如果有其他节点，则Merkle Root是根节点哈希值和其他节点哈希值的哈希值  
	hasher := sha256.New()  
	hasher.Write([]byte(rootHash))  
	for _, h := range hashes {  
		hasher.Write([]byte(h))  
	}  
	merkleRoot := hex.EncodeToString(hasher.Sum(nil))  
  
	return merkleRoot, nil  
}  
  
// ErrUnknownNodeType 是当遇到未知节点类型时返回的错误。  
var ErrUnknownNodeType = error("unknown node type")

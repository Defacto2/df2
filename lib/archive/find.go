package archive

import (
	"path/filepath"
	"strings"
)

func FindNFO(name string, files ...string) string {
	f := make(finds)
	for _, file := range files {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		fn := strings.ToLower(file)
		ext := filepath.Ext(fn)
		switch ext {
		case diz, nfo, txt:
			// okay
		default:
			continue
		}
		switch {
		case fn == base+".nfo": // [archive name].nfo
			f[file] = 1
		case fn == base+".txt": // [archive name].txt
			f[file] = 2
		case ext == ".nfo": // [random].nfo
			f[file] = 3
		case fn == "file_id.diz": // BBS file description
			f[file] = 4
		case fn == base+".diz": // [archive name].diz
			f[file] = 5
		case fn == ".txt": // [random].txt
			f[file] = 6
		case fn == ".diz": // [random].diz
			f[file] = 7
		default: // currently lacking is [group name].nfo and [group name].txt priorities
		}
	}
	return f.top()
}

package services

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/pkg/bookparser"
)

// FSService provides native OS file dialogs and filesystem helpers for Wails v3.
type FSService struct{}

// NewFSService creates a new FSService instance.
func NewFSService() *FSService {
	return &FSService{}
}

// PickBookFile opens a native OS file dialog to select a single novel or ebook file.
func (f *FSService) PickBookFile() string {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseFiles(true).
		ResolvesAliases(true)

	if runtime.GOOS == "darwin" {
		dialog.SetMessage("Select Novel or E-Book File")
	} else {
		dialog.SetTitle("Select Novel or E-Book File")
	}

	dialog.AddFilter("All Supported Formats (*.epub, *.mobi, *.azw3, *.txt, *.md, *.docx, *.pdf, *.fb2, *.cbz...)", "*.epub;*.kepub.epub;*.mobi;*.azw;*.azw3;*.amz;*.fb2;*.fbz;*.txt;*.md;*.markdown;*.text;*.doc;*.docx;*.odt;*.rtf;*.pdf;*.html;*.htm;*.xhtml;*.cbz;*.cbr;*.cbt;*.cb7;*.zip;*.rar;*.7z;*.m4b;*.mp3;*.m4a;*.flac;*.ogg;*.wav;*.aac;*.csv;*.tsv;*.tex;*.latex;*.pptx;*.ppt;*.xlsx;*.xls")
	dialog.AddFilter("EPUB & Kindle E-Books (*.epub, *.mobi, *.azw, *.azw3)", "*.epub;*.kepub.epub;*.mobi;*.azw;*.azw3;*.amz")
	dialog.AddFilter("Plain Text & Markdown (*.txt, *.md)", "*.txt;*.md;*.markdown;*.text")
	dialog.AddFilter("Office & Document Files (*.docx, *.doc, *.pdf, *.odt, *.rtf)", "*.docx;*.doc;*.pdf;*.odt;*.rtf")
	dialog.AddFilter("FictionBook Files (*.fb2, *.fbz)", "*.fb2;*.fbz")
	dialog.AddFilter("Comic Books & Archives (*.cbz, *.cbr, *.zip, *.rar, *.7z)", "*.cbz;*.cbr;*.cbt;*.cb7;*.zip;*.rar;*.7z")
	dialog.AddFilter("Web Pages (*.html, *.htm)", "*.html;*.htm;*.xhtml")
	dialog.AddFilter("Audiobooks (*.m4b, *.mp3, *.m4a)", "*.m4b;*.mp3;*.m4a;*.flac;*.ogg;*.wav;*.aac")
	dialog.AddFilter("All Files (*.*)", "*.*")

	if path, err := dialog.PromptForSingleSelection(); err == nil {
		return path
	}
	return ""
}

// PickBookFiles opens a native OS file dialog to select multiple novel or ebook files.
func (f *FSService) PickBookFiles() []string {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseFiles(true).
		ResolvesAliases(true)

	if runtime.GOOS == "darwin" {
		dialog.SetMessage("Select Novel or E-Book Files")
	} else {
		dialog.SetTitle("Select Novel or E-Book Files")
	}

	dialog.AddFilter("All Supported Formats (*.epub, *.mobi, *.azw3, *.txt, *.md, *.docx, *.pdf, *.fb2, *.cbz...)", "*.epub;*.kepub.epub;*.mobi;*.azw;*.azw3;*.amz;*.fb2;*.fbz;*.txt;*.md;*.markdown;*.text;*.doc;*.docx;*.odt;*.rtf;*.pdf;*.html;*.htm;*.xhtml;*.cbz;*.cbr;*.cbt;*.cb7;*.zip;*.rar;*.7z;*.m4b;*.mp3;*.m4a;*.flac;*.ogg;*.wav;*.aac;*.csv;*.tsv;*.tex;*.latex;*.pptx;*.ppt;*.xlsx;*.xls")
	dialog.AddFilter("EPUB & Kindle E-Books (*.epub, *.mobi, *.azw, *.azw3)", "*.epub;*.kepub.epub;*.mobi;*.azw;*.azw3;*.amz")
	dialog.AddFilter("Plain Text & Markdown (*.txt, *.md)", "*.txt;*.md;*.markdown;*.text")
	dialog.AddFilter("Office & Document Files (*.docx, *.doc, *.pdf, *.odt, *.rtf)", "*.docx;*.doc;*.pdf;*.odt;*.rtf")
	dialog.AddFilter("FictionBook Files (*.fb2, *.fbz)", "*.fb2;*.fbz")
	dialog.AddFilter("Comic Books & Archives (*.cbz, *.cbr, *.zip, *.rar, *.7z)", "*.cbz;*.cbr;*.cbt;*.cb7;*.zip;*.rar;*.7z")
	dialog.AddFilter("Web Pages (*.html, *.htm)", "*.html;*.htm;*.xhtml")
	dialog.AddFilter("Audiobooks (*.m4b, *.mp3, *.m4a)", "*.m4b;*.mp3;*.m4a;*.flac;*.ogg;*.wav;*.aac")
	dialog.AddFilter("All Files (*.*)", "*.*")

	if paths, err := dialog.PromptForMultipleSelection(); err == nil {
		return paths
	}
	return []string{}
}

// PickFolder opens a native OS folder picker dialog.
func (f *FSService) PickFolder() string {
	dialog := application.Get().Dialog.OpenFile().
		CanChooseDirectories(true).
		CanCreateDirectories(true).
		ResolvesAliases(true)

	if runtime.GOOS == "darwin" {
		dialog.SetMessage("Select Book Directory")
	} else {
		dialog.SetTitle("Select Book Directory")
	}

	if path, err := dialog.PromptForSingleSelection(); err == nil {
		return path
	}
	return ""
}

// ScanFolderForBooks finds all supported ebook and novel files in a given directory.
func (f *FSService) ScanFolderForBooks(dirPath string) []string {
	if dirPath == "" {
		return []string{}
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return []string{}
	}
	var books []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if bookparser.IsAllowedBookFormat(entry.Name()) {
			books = append(books, filepath.Join(dirPath, entry.Name()))
		}
	}
	sort.Strings(books)
	return books
}

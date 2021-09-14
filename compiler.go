package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var CACHED_HTML = make(map[string]string)

// Copy copies a src file to dst
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// CopyDirectory recursively copies the contents of a directory
func CopyDirectory(src, dst string) error {
	srcDir, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcDir.Close()

	ls, err := srcDir.Readdir(-1)
	if err != nil {
		return err
	}

	srcDirInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcDirInfo.Mode())
	if err != nil {
		log.Printf("ERROR: could not make dir '%s' with error '%s'\n", dst, err)
	}

	for _, entry := range ls {
		srcName := path.Join(src, entry.Name())
		dstName := path.Join(dst, entry.Name())
		if entry.IsDir() {
			log.Println("attempting to make directory " + dstName)
			err = os.MkdirAll(dstName, entry.Mode())
			if err != nil {
				log.Printf("ERROR: could not make dir '%s' with error '%s'\n", dstName, err)
			}

			err = CopyDirectory(srcName, dstName)
			if err != nil {
				log.Printf("ERROR: could not copy '%s' to '%s' with error '%s'\n", srcName, dstName, err)
			}
		} else {
			err = Copy(srcName, dstName)
			if err != nil {
				log.Printf("ERROR: could not copy '%s' to '%s' with error '%s'\n", srcName, dstName, err)
			}
		}
	}

	return nil
}

// PrettifyHtml prettifies the HTML in `f` and writes it to `output/fname`.
func PrettifyHtml(f io.Reader, fname, outputDir string) {
	err := os.MkdirAll(outputDir, 0777)
	if err != nil {
		log.Fatalf("ERROR: Could not prettify html, failed to create output directory with error: %s", err)
	}

	// parse the actual file name based on the last slash in the path
	lastSlash := strings.LastIndex(fname, "/")
	if lastSlash < 0 {
		lastSlash = 0
	}
	fileName := fname[lastSlash:]

	// write to a tmp file first
	outputFile, err := os.Create(path.Join(outputDir, fileName+".tmp"))
	if err != nil {
		log.Fatalf("ERROR: Could not prettify html, failed to create output file with error: %s", err)
	}
	defer outputFile.Close()

	if err != nil {
		log.Fatalf("ERROR: Could not prettify html, failed to create output file with error: %s", err)
	}

	indentationLevel := 0
	currentIndentationLevel := 0

	tags := make(Stack, 0)

	z := html.NewTokenizer(f)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				// check for unclosed tags
				if !tags.IsEmpty() {
					t := tags.Pop()
					log.Println("ERROR: Unclosed tag: ", t, ". Output may be inconsistent.")
				}
				// we're done prettifying! rename tmp file to the appropriate name
				err = os.Rename(path.Join(outputDir, fileName+".tmp"), path.Join(outputDir, fileName))

				if err != nil {
					log.Fatalf("ERROR: Failed to prettify '%s' with error: %s", fname, err)
				}

				return
			}
			log.Fatalf("ERROR: Failed to prettify html with error: %s", err)
		}

		tok := z.Token()

		if strings.TrimSpace(tok.String()) != "" {

			if tt == html.StartTagToken && tt != html.EndTagToken {
				tags.Push(tok.DataAtom)
				indentationLevel++
			} else if tt == html.EndTagToken && tok.DataAtom == tags.Peek() {
				tags.Pop()
				indentationLevel--
				currentIndentationLevel = indentationLevel
			}

			for i := 0; i < currentIndentationLevel; i++ {
				fmt.Fprint(outputFile, "    ")
			}

			if tt != html.TextToken {
				fmt.Fprintln(outputFile, strings.TrimSpace(tok.String()))
			} else {
				fmt.Fprintln(outputFile, strings.TrimSpace(tok.Data))
			}

			currentIndentationLevel = indentationLevel
		}
	}

}

// readTagName returns the tag name from a tag
func readTagName(tag string) string {
	endIndex := strings.Index(tag, "/>")
	if endIndex < 0 {
		endIndex = len(tag) - 1
	}
	return strings.TrimSpace(tag[1:endIndex])
}

// getComponents gets the component tags from a line of text
func getComponents(line string) []string {
	components := make([]string, 0)
	for i := 0; i < len(line); i++ {
		if rune(line[i]) == '<' {
			tag := ""
			if i+1 < len(line) {
				if rune(line[i+1]) != '/' {
					tag = readTagName(line[i:])
					i++
					if strings.HasPrefix(tag, "app-") {
						components = append(components, tag[4:])
					}
				}
			}
		}
	}
	return components
}

// validateComponents makes sure there are no circular dependencies
func validateComponents(directory string) error {
	component_dir := directory + "/components"
	valid := true

	dependencies := make(map[string]*Set)

	// gather dependencies
	files, err := ioutil.ReadDir(component_dir)
	if err != nil {
		return err
	}

	for _, component := range files {
		if !component.IsDir() {
			name := component.Name()
			name = name[:len(name)-5]
			_, inMap := dependencies[name]

			if strings.HasSuffix(component.Name(), ".html") && !inMap {
				dependencies[name] = NewSet()

				componentFile, err := os.Open(component_dir + "/" + component.Name())
				if err != nil {
					return err
				}
				defer componentFile.Close()

				scanner := bufio.NewScanner(componentFile)
				for scanner.Scan() {
					for _, dep := range getComponents(scanner.Text()) {
						dependencies[name].Add(dep)
					}
				}
			}
		}
	}

	// check for circular dependencies
	for component, deps := range dependencies {
		componentDependencies := deps.ShallowCopy()

		for _, d := range deps.Values() {
			d_dependencies := dependencies[d.(string)].ShallowCopy()

			componentDependencies.Update(d_dependencies.Values())
		}

		if componentDependencies.Contains(component) {
			log.Printf("FATAL ERROR: component '%s' is dependent on itself\n", component)
			valid = false
		}
	}

	if !valid {
		return errors.New("could not compile: one or more components have circular dependencies")
	}

	return nil
}

// getComponentHtml retrieves a components html, recursively replacing any components within it
func getComponentHtml(directory, component string, lineNumber int) string {
	html, cached := CACHED_HTML[component]

	if cached {
		return html
	}

	componentHtml, err := os.Open(path.Join(directory, "components", component+".html"))
	if err != nil {
		log.Printf("WARNING:\t component '%s' found on line %d but no '%s.html' was found in '%s'", component, lineNumber, component, path.Join(directory, "components/"))
		return ""
	}

	scanner := bufio.NewScanner(componentHtml)
	html = ""
	for lineNumberScanned := 0; scanner.Scan(); lineNumberScanned++ {
		html += processLine(directory, scanner.Text(), lineNumberScanned)
	}

	CACHED_HTML[component] = html
	return html
}

// processLine replaces component tags with components
func processLine(directory, line string, lineNumber int) string {
	retval := line
	components := getComponents(line)

	if len(components) > 0 {
		retval = ""
		for _, component := range components {
			componentTag := "<app-" + component + "/>"
			componentIndex := strings.Index(line, componentTag)
			if componentIndex < 0 {
				componentIndex = len(line) - 1
			}
			retval += line[:componentIndex] + getComponentHtml(directory, component, lineNumber)
			line = line[componentIndex+len(componentTag):]
		}
	}
	return retval
}

// stripComments strips html comments from a string
func stripComments(html string) string {
	return regexp.MustCompile("(?s)(<!--.*?-->)").ReplaceAllString(html, "")
}

func main() {
	// flag parsing
	var outputDir string
	var directory string
	var prettify string

	flag.StringVar(&prettify, "prettify", "", "File to prettify, output to directory indicated by --out flag")
	flag.StringVar(&directory, "in", ".", "Directory of html to compile (use '/' not '\\')")
	flag.StringVar(&outputDir, "out", "", "Directory to write output to (use '/' not '\\')")
	flag.Parse()

	if outputDir == "" {
		outputDir = path.Join(directory, "output")
	}

	// file to process
	fname := prettify
	if fname == "" {
		fname = path.Join(directory, "index.html")
	}

	// open file to process
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("ERROR: Could not open file '%s' with error: %s\n", fname, err)
	}
	defer f.Close()

	// if we just need to prettify something, just prettify it and return
	if prettify != "" {
		log.Printf("Prettifying `%s` to output directory `%s`\n", fname, outputDir)
		PrettifyHtml(f, fname, outputDir)
		return
	}

	// compile the html components
	log.Printf("Compiling '%s'...\n", directory)

	// check for circular dependencies
	err = validateComponents(directory)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	// strip comments from html
	indexHtml, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	strippedOfComments := stripComments(string(indexHtml))

	// process html and compile
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(strippedOfComments))
	for lineNumber := 0; scanner.Scan(); lineNumber++ {
		output += processLine(directory, scanner.Text(), lineNumber)
	}

	// prettify it
	PrettifyHtml(strings.NewReader(output), "index.html", outputDir)

	// copy .js .css and .html files to output directory
	files, _ := ioutil.ReadDir(directory)
	for _, file := range files {
		srcName := path.Join(directory, file.Name())
		dstName := path.Join(outputDir, file.Name())
		if file.IsDir() && (strings.Contains(file.Name(), "js") || strings.Contains(file.Name(), "css") || strings.Contains(file.Name(), "html")) {
			err = CopyDirectory(srcName, dstName)
			if err != nil {
				log.Printf("ERROR: could not copy '%s' to '%s' with error '%s'\n", srcName, dstName, err)
			}
		}
		if strings.HasSuffix(file.Name(), ".js") || strings.HasSuffix(file.Name(), ".css") || strings.HasSuffix(file.Name(), ".html") && file.Name() != "index.html" {
			err = Copy(srcName, dstName)
			if err != nil {
				log.Printf("ERROR: could not copy '%s' to '%s' with error '%s'\n", srcName, dstName, err)
			}
		}
	}

	log.Println("Done!")
}

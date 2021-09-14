# html-compile
`html-compile` is a dead simple HTML component translating compiler that lets you write modular components, but in vanilla HTML / js / css! 

If it finds a self-closing tag with the label "app-", it attempts to replace that tag with the HTML file with the same name. Comments are not included in the compiled file.

`html-compile` will copy all HTML, CSS, and JS files from the input directory, and recursively copy all directories in the input directory named "html", "js", or "css" to the output directory.

`html-compile` can also be used to prettify existing HTML, but won't fix HTML errors, just make existing HTML have consistent indentation. 

`html-compile` follows the simple philosophy of do two things and do them good enough.

## How to use 
1. Install the cli: `go install github.com/reesporte/html-compile@latest`

2. Set up a directory with an `index.html` file.

3. Then, to create a component, make a `COMPONENT.html` file and put it in a `components` directory which is on the same level as your `index.html`. Then, to use your component, put the tag `<app-COMPONENT/>` in your `index.html` file where you want to use the component. 

4. To compile your HTML, just run `html-compile --in <directory of index.html>`.

To prettify existing HTML inplace, run `html-compile --prettify <html file to prettify> --out <directory of file>`. So if you were prettifying a file in the current directory, you could run `html-compile --prettify file.html --out .`.



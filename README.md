# html-compile
`html-compile` is a dead simple HTML component translating compiler that lets you write modular components, but in vanilla HTML / JS / CSS!

## Installation
```bash
go get -u github.com/reesporte/html-compile
go install github.com/reesporte/html-compile@latest
```

## What it does 
If it finds a self-closing tag with the label "app-", it attempts to replace that tag with the HTML file with the same name. Comments are not included in the compiled file. It will copy all HTML, CSS, and JS files from the input directory, and recursively copy all directories in the input directory named "html", "js", or "css" to the output directory. 

It can also be used to prettify HTML, but won't fix HTML errors, just make existing HTML have consistent indentation. 


### Component transpiling:
1. Set up a directory with an `index.html` file.
2. Then, to create a component, make a `COMPONENT.html` file and put it in a `components` directory which is on the same level as your `index.html`. Then, to use your component, put the tag `<app-COMPONENT/>` in your `index.html` file where you want to use the component. 
3. To compile your HTML, just run `html-compile --in <directory of index.html>`.

### Prettify HTML inplace:
```bash
html-compile --prettify <html file to prettify> --out <directory of file>
``` 

### Prettify a file in the current directory:
```bash
html-compile --prettify file.html --out .
```


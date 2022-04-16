# PDFtoJPEG

## usage:
run `go build` from main directory, then launch PDFConverter  
visit API at localhost:8080 and upload files.  
wait for response, which will be a .zip of the extracted images, organized in directories named after the input file.

converter.py can be used as a standalone tool to convert single documents.  
it requires and input file and an output directory to creae its own output directory inside of.
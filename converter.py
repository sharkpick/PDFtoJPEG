#!/usr/bin/env python3
from pdf2image import convert_from_path
import sys
import os


def processFile(inputFile, outputDirectory):
    dirname = os.path.dirname(inputFile)
    base = os.path.splitext(os.path.basename(inputFile))[0]
    newDir = outputDirectory
    if newDir[-1] != "/":
        newDir += "/"
    newDir += base
    try:
        os.makedirs(newDir)
    except FileExistsError:
        pass
    images = convert_from_path(inputFile, dpi=300)
    for i in range(len(images)):
        images[i].save(newDir + "/page"+str(i)+".jpg", "JPEG")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        sys.stderr.write("Error: requires input file")
        sys.exit(99)
    processFile(sys.argv[1], sys.argv[2])
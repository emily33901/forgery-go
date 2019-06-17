# Forgery

A hopefully 1:1 replacement for the veteran Hammer World Editor by Valve Corporation.

Currently supporting viewing maps in 3d with both a wireframe and textured views (2D in the works but not quite there yet).

A large amount of the code here is from Galaco's work on [Lambda](https://github.com/Galaco/Lambda) & [Lambda-Client](https://github.com/Galaco/Lambda-Client). 

## Ideas list

VPKs could (should...) be memory mapped into the process - this would allow us to just stream out texture data without having to
copy it into memory

Ideally something like the IDataCache & IMDLCache could be implemented to allow for better accessing of model and general data.
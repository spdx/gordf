# gordf
[![Build Status](https://github.com/RishabhBhatnagar/goRdf/workflows/go%20test/badge.svg)](https://github.com/spdx/tools-golang/actions)  [![Coverage Status](https://coveralls.io/repos/github/RishabhBhatnagar/gordf/badge.svg?branch=master)](https://coveralls.io/github/RishabhBhatnagar/gordf?branch=master)


gordf is a package which provides a parser for RDF files linearized using RDF/XML format. It will be used to represent the rdf files in memory and write back in possibly different formats like json, and xml.

# License
[MIT](https://github.com/spdx/goRdf/blob/master/LICENSE.txt)

Note: the license text in the [sample-docs input file](examples/sample-docs/input.rdf) does not apply to any of the files (including the sample) - it is for illustrative purposes on how license text can be represented in SPDX RDF format.

# Requirements
At present, gordf does not require any addittional packages except the base library packages of golang.

# Installation Guide
  Make sure that the `GOPATH` and `GOROOT` variables is correctly set in the current environment.
  Recommended go-version is go1.4
 * Using GoLang's Package Manager 
      <pre>go get github.com/spdx/gordf</pre>

 * Directly From GitHub
      <pre> git clone github.com/spdx/gordf gordf </pre>
 
 * Getting a Specific Version
      <pre> go get -u github.com/spdx/gordf@vx.y.z </pre>
      `vx.y.z` is a valid version from the tags of this repository.
      
 * Updating the Local Installation
      <pre> go get -u github.com/spdx/gordf </pre>


# Development Status 
The repository is in its preliminary stage of development. It might have some defects. For reporting any issue, you can raise a ticket [here](https://github.com/spdx/goRdf/issues).

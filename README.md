# gotter


A go command line tool to help you manage your go projects into a single and consistent workspace.

### If you do the following...You probably need gotter
- Have you Go projects in a seperate folder, and the rest of your projects in a 'workspace' folder because Go code needs to be in a valid `$GOPATH`
- Like having your Go projects together in single folder, not in `$GOPATH/github.com/jsimnz/gotter`
- Are always battling between wanting to `git clone` your repo or `go get`
- Move your code into a 'workspace' folder after `go get` has put it in an ackward `$GOPATH` location
- Always updating the `remote origin` of your packages that `go get` download as `https://`
- Like trying new tools :smile:

## How it works

It uses the same syntax as `go get` so you should already be firmiliar with it. When you call `gotter get` it uses the go toolchain to download your package/repo into it's appropriate fully quantified name folder. Then creates a symlink between that and a defined `$WORKSPACE`, and finally (TODO), if possible updates the `git remote origin` of the package to use ssh so you can use your public key authentication.

## Usage

#### download, link, and update origin URL of a Go package
```
gotter get github.com/jsimnz/gotter
```

This will download your package using the go tool chain, create a symlink from your package folder in your `$GOPATH` and will update the remote origin URL.

Notice that it uses the same syntax as `go get` for the URL. The package FQN. But it also supports other git URLs such as `git://URL`. `http(s)://URL`. and even `git@URL:repo.git`. If you use the latter URLs and not the FQN, It will keep the remote origin URL as is. 

#### Just download
```
gotter clone github.com/jsimnz/gotter
```
This will use the go tool chain to download your package, just like before, but won't create a symlink to your `$WORKSPACE`, and won't update the origin URL.

#### Link the package
```
gotter link github.com/jsimnz/gotter
```

This will create a symlink from your $GOPATH/project to your $WORKSPACE/project

#### For help
```
gotter --help OR gotter -h
```

## Install

You can simply `go get github.com/jsimnz/gotter`

## TODO
- Finish remote origin URL update
- Use config file to set `$WORKSPACE` and other settings
- Expose remote origin as a sub command

## License

The MIT License (MIT)

Copyright (c) 2014 John-Alan Simmons

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
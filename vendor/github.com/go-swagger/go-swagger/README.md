# Swagger 2.0 [![Build Status](https://circleci.com/gh/go-swagger/go-swagger.svg?style=shield)](https://circleci.com/gh/go-swagger/go-swagger) [![Build status](https://ci.appveyor.com/api/projects/status/x377t5o9ennm847o/branch/master?svg=true)](https://ci.appveyor.com/project/casualjim/go-swagger/branch/master) [![codecov](https://codecov.io/gh/go-swagger/go-swagger/branch/master/graph/badge.svg)](https://codecov.io/gh/go-swagger/go-swagger) [![Slack Status](https://slackin.goswagger.io/badge.svg)](https://slackin.goswagger.io)

[![license](http://img.shields.io/badge/license-Apache%20v2-orange.svg)](https://raw.githubusercontent.com/swagger-api/swagger-spec/master/LICENSE) [![GoDoc](https://godoc.org/github.com/go-swagger/go-swagger?status.svg)](http://godoc.org/github.com/go-swagger/go-swagger) [![GitHub version](https://badge.fury.io/gh/go-swagger%2Fgo-swagger.svg)](https://badge.fury.io/gh/go-swagger%2Fgo-swagger) [![Docker Repository on Quay](https://quay.io/repository/goswagger/swagger/status "Docker Repository on Quay")](https://quay.io/repository/goswagger/swagger)

Development of this toolkit is sponsored by VMware:<br>[![VMWare](https://avatars2.githubusercontent.com/u/473334?v=3&s=200)](https://vmware.github.io)  

Contains an implementation of Swagger 2.0. It knows how to serialize and deserialize swagger specifications.

Swagger is a simple yet powerful representation of your RESTful API.<br>With the largest ecosystem of API tooling on the planet, thousands of developers are supporting Swagger in almost every modern programming language and deployment environment.

With a Swagger-enabled API, you get interactive documentation, client SDK generation and discoverability. We created Swagger to help fulfill the promise of APIs.

Swagger helps companies like Apigee, Getty Images, Intuit, LivingSocial, McKesson, Microsoft, Morningstar, and PayPal build the best possible services with RESTful APIs. Now in version 2.0, Swagger is more enabling than ever. And it's 100% open source software.

## How is this different from go generator in swagger-codegen

tl;dr The main difference at this moment is that this one will actually work.

The swagger-codegen project only generates a client and even there it will only support flat models.

* This project supports most features offered by jsonschema including polymorphism.
* It allows for generating a swagger specification from go code.
* It allows for generating a server from a swagger definition and to generate an equivalent spec back from that codebase.
* It allows for generating a client from a swagger definition.
* It has support for several common swagger vendor extensions.

Why is this not done in the swagger-codegen project? Because:

* I don't really know java very well and so I'd be learning both java and the object model of the codegen which was in heavy flux as opposed to doing go and I really wanted to go experience of designing a large codebase with it.
* Go's super limited type system makes it so that it doesn't fit well in the model of swagger-codegen
* Go's idea of polymorphism doesn't reconcile very well with a solution designed for languages that actually have inheritance and so forth.
* For supporting types like [][][]map[string][][]int64 I don't think it's possible with mustache
* I gravely underestimated the amount of work that would be involved in making something useful out of it.
* My personal mission: I want the jvm to go away, it was great way back when now it's just silly (vm in container on vm in vm in container)

## Using 0.5.0

Because 0.5.0 and master have diverged significantly, you should checkout the tag 0.5.0 for go-swagger when you use the currently released version.

## Migrating

### From 0.5.0 to 0.6.0

You will have to rename some imports:

```
github.com/go-swagger/go-swagger/httpkit/validate to github.com/go-openapi/validate
github.com/go-swagger/go-swagger/httpkit to github.com/go-openapi/runtime
github.com/naoina/denco to github.com/go-openapi/runtime/middleware/denco
github.com/go-swagger/go-swagger to github.com/go-openapi
```

## Docs

<https://goswagger.io>

### Binary distribution

go-swagger is distributed as binaries that are built of signed tags. It is published as github release, rpm, deb and docker image.

#### Docker image

```shell
docker pull quay.io/goswagger/swagger

alias swagger="docker run --rm -it -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger"
swagger version
```

#### Homebrew/Linuxbrew

```shell
brew tap go-swagger/go-swagger
brew install go-swagger
```

#### Static binary

You can download a binary for your platform from github:

<https://github.com/go-swagger/go-swagger/releases/latest>

```shell
latestv=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/latest | jq -r .tag_name)
curl -o /usr/local/bin/swagger -L'#' https://github.com/go-swagger/go-swagger/releases/download/$latestv/swagger_$(echo `uname`|tr '[:upper:]' '[:lower:]')_amd64
chmod +x /usr/local/bin/swagger
```

#### Debian packages [ ![Download](https://api.bintray.com/packages/go-swagger/goswagger-debian/swagger/images/download.svg) ](https://bintray.com/go-swagger/goswagger-debian/swagger/_latestVersion)

This repo will work for any debian, the only file it contains gets copied to /usr/bin

```shell
echo "deb https://dl.bintray.com/go-swagger/goswagger-debian ubuntu main" | sudo tee -a /etc/apt/sources.list
```

#### RPM packages [ ![Download](https://api.bintray.com/packages/go-swagger/goswagger-rpm/swagger/images/download.svg) ](https://bintray.com/go-swagger/goswagger-rpm/swagger/_latestVersion)

This repo should work on any distro that wants rpm packages, the only file it contains gets copied to /usr/bin/

```shell
wget https://bintray.com/go-swagger/goswagger-rpm/rpm -O bintray-go-swagger-goswagger-rpm.repo
```

### From source

Install or update from source:

```
go get -u github.com/go-swagger/go-swagger/cmd/swagger
```

The implementation also provides a number of command line tools to help working with swagger.

Currently there is a [spec validator tool](https://goswagger.io/usage/validate.html):

```
swagger validate https://raw.githubusercontent.com/swagger-api/swagger-spec/master/examples/v2.0/json/petstore-expanded.json
```

To generate a [server for a swagger spec](https://goswagger.io/generate/server.html) document:

```
swagger generate server [-f ./swagger.json] -A [application-name [--principal [principal-name]]
```

To generate a [client for a swagger spec](https://goswagger.io/generate/client.html) document:

```
swagger generate client [-f ./swagger.json] -A [application-name [--principal [principal-name]]
```

To generate a [swagger spec document for a go application](https://goswagger.io/generate/spec.html):

```
swagger generate spec -o ./swagger.json
```

### Bash Completion

Bash completion is supported and can be activated as follows:

```
source ./cmd/swagger/completion/swagger.bash-completion
```

Note that this does require you already setup bash completion,
which can be done in 2 simple steps:

1) install `bash-completion` using your favourite package manager;

2) run `source /etc/bash_completion` in bash;

### Zsh Completion

Zsh completion is supported and can be copied/soft-linked from:

```
./cmd/swagger/completion/swagger.zsh-completion
```

In case you're new to adding auto-completion to zsh completion,
here is how you could enable swagger's zsh completion step by step:

1) create a folder used to store your completions (eg. `$HOME/.zsh/completion`);

2) append the following to your `$HOME/.zshrc` file:

```
# add auto-completion directory to zsh's fpath
fpath=($HOME/.zsh/completion $fpath)

# compsys initiatlization
autoload -U compinit
compinit
```

3) copy/soft-link `./cmd/swagger/completion/swagger.zsh-completion` to `$HOME/.zsh/completion/_swagger`;

## Licensing

The toolkit itself is licensed as Apache Software License 2.0. Just like swagger, this does not cover code generated by the toolkit. That code is entirely yours to license however you see fit.

## What's inside?

For a V1 I want to have this feature set completed:

- [x] Documentation site
- [x] Play nice with golint, go vet etc.
- [x] An object model that serializes to swagger yaml or json
- [x] A tool to work with swagger:

  - [x] validate a swagger spec document:
  - [x] validate against jsonschema
  - [ ] validate extra rules outlined [here](https://github.com/apigee-127/sway/blob/master/docs/versions/2.0.md#semantic-validation)

    - [x] definition can't declare a property that's already defined by one of its ancestors (Error)
    - [x] definition's ancestor can't be a descendant of the same model (Error)
    - [x] each api path should be non-verbatim (account for path param names) unique per method (Error)
    - [ ] each security reference should contain only unique scopes (Warning)
    - [x] each path parameter should correspond to a parameter placeholder and vice versa (Error)
    - [x] path parameter declarations do not allow empty names _(`/path/{}` is not valid)_ (Error)
    - [x] each definition property listed in the required array must be defined in the properties of the model (Error)
    - [x] each parameter should have a unique `name` and `in` combination (Error)
    - [x] each operation should have at most 1 parameter of type body (Error)
    - [x] each operation cannot have both a body parameter and a formData parameter (Error)
    - [x] each operation must have an unique `operationId` (Error)
    - [x] each reference must point to a valid object (Error)
    - [x] each referencable definition must have references (Warning)
    - [x] every default value that is specified must validate against the schema for that property (Error)
    - [x] every example that is specified must validate against the schema for that property (Error)
    - [x] items property is required for all schemas/definitions of type `array` (Error)

  - [x] serve swagger UI for any swagger spec file
  - [x] code generation
  - [x] generate api based on swagger spec
  - [x] generate go client from a swagger spec
  - [x] spec generation
  - [x] generate spec document based on the code

    - [x] generate meta data (top level swagger properties) from package docs
    - [x] generate definition entries for models
    - [x] support composed structs out of several embeds
    - [x] support allOf for composed structs
    - [x] generate path entries for routes
    - [x] generate responses from structs
    - [x] support composed structs out of several embeds
    - [x] generate parameters from structs
    - [x] support composed structs out of several embeds

- [x] Middlewares:

  - [x] serve spec
  - [x] routing
  - [x] validation
  - [x] additional validation through an interface
  - [x] authorization

    - [x] basic auth
    - [x] api key auth
    - [x] oauth2 bearer auth

  - [x] swagger docs UI

- [x] Typed JSON Schema implementation

  - [x] JSON Pointer that knows about structs
  - [x] JSON Reference that knows about structs
  - [x] Passes current json schema test suite

- [x] extended string formats

  - [x] uuid, uuid3, uuid4, uuid5
  - [x] email
  - [x] uri (absolute)
  - [x] hostname
  - [x] ipv4
  - [x] ipv6
  - [x] mac
  - [x] credit card
  - [x] isbn, isbn10, isbn13
  - [x] social security number
  - [x] hexcolor
  - [x] rgbcolor
  - [x] date
  - [x] date-time
  - [x] duration
  - [x] password
  - [x] custom string formats

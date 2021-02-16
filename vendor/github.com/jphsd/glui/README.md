# Simple Image Display For Go

This package wraps the [Go port](https://pkg.go.dev/github.com/go-gl/glfw/v3.3/glfw) of the [C implementation of GLFW](https://www.glfw.org/documentation.html) in some boilerplate to make displaying an image on the screen almost trivial.

The version of [GL used is 4.1](https://pkg.go.dev/github.com/go-gl/gl/v4.1-core/gl) and the boilerplate includes the necessary shaders to perform the rendering.

See glui/cmd/image.go for a simple example of reading an image and displaying it in a window.

	$ go run image.go <your image here - most image types supported>

The glui/cmd/images.go version takes multiple images on the command line and rotates through them. This example also demonstrates attaching a call back for character processing - hitting ESC will close the window.

Functions are included to allow you to create new windows with a backing image, set that image to some other image and a loop function (which actually does the rendering). The loop function can take a zero argument function that will be called once per loop iteration. In addition to handling the window rendering, the loop function also handles any call backs registered on the windows.

Finally, the events.c program provided in the C implementation of GLFW has been ported over to Go and is in glui/cmd/events.go.

	$ go run events.go

This will dump out all the event types, for which call backs can be registered, to the terminal. It doesn't utilize the glui package itself.

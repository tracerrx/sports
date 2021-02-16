// Package glui provides a simple way to display an image or images on a screen, each in its own window.
// The package makes use of the GLFW library and OpenGL v4.1
package glui

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"
	"strings"
	"sync"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// GLWin ties an image to a window
type GLWin struct {
	Img   *image.RGBA
	Win   *glfw.Window
	dirty bool
	lock  sync.Mutex
}

// WinMap provides a map of all windows to their corresponding GLWin
var WinMap map[*glfw.Window]*GLWin = make(map[*glfw.Window]*GLWin)
var initialized bool = false

func init() {
	// GLFW event handling *must* run on the main OS thread
	runtime.LockOSThread()
}

// NewGLWin is used to create a new window of width w and height h. If the window is decorated
// then the title will appear in the window frame. Note - the window and image sizes are independent
// of each other. The image is scaled (not cropped) to fit the window.
func NewGLWin(w, h int, title string, img image.Image, decorated bool) *GLWin {
	if !initialized {
		// Inititalize GLFW package
		if err := glfw.Init(); err != nil {
			panic(err)
		}
		// Context creation hints
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	}

	if decorated {
		glfw.WindowHint(glfw.Decorated, glfw.True)
	} else {
		glfw.WindowHint(glfw.Decorated, glfw.False)
	}
	window, err := glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		panic(err)
	}

	if !initialized {
		window.MakeContextCurrent()
		glInitialize()
		initialized = true
	}

	glwin := &GLWin{convertImage(img), window, true, sync.Mutex{}}
	WinMap[window] = glwin

	// Set window refresh callback
	window.SetRefreshCallback(func(win *glfw.Window) {
		glfwRender(glwin)
	})

	return glwin
}

func convertImage(img image.Image) *image.RGBA {
	// Convert to RGBA if not already
	rgba, ok := img.(*image.RGBA)
	if !ok {
		// Perform color model conversion
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
	}
	// All bytes need to be contiguous
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic(fmt.Errorf("unsupported stride"))
	}
	return rgba
}

// SetImage is used to change the current image in a window. It's thread safe.
func (w *GLWin) SetImage(img image.Image) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.Img = convertImage(img)
	w.dirty = true
}

// Loop is how window events get processed and the images rendered to their windows. It will run
// until there are no windows left to process. The update function provides a way to insert code
// into this loop - it should be non-blocking otherwise window updates will stall.
func Loop(update func()) {
	defer glfw.Terminate()

	glfw.SwapInterval(1) // Wait for next vsync

	for len(WinMap) > 0 {
		// Perform window maintenance
		for window, glwin := range WinMap {
			if window.ShouldClose() {
				window.Destroy()
				delete(WinMap, window)
			}
			if glwin.dirty {
				glfwRender(glwin)
				glwin.dirty = false
			}
		}
		// Run user update function
		if update != nil {
			update()
		}
		// Check for GLFW callbacks
		glfw.PollEvents()
	}
}

func glfwRender(win *GLWin) {
	window := win.Win
	window.MakeContextCurrent()

	// Not required since the image maps to the entire window
	//	gl.ClearColor(1.0, 0.0, 0.0, 1.0) // Red!
	//	gl.Clear(gl.COLOR_BUFFER_BIT)

	// Render
	gl.UseProgram(program)
	gl.BindVertexArray(vao)
	gl.ActiveTexture(gl.TEXTURE0)
	// Load the texture
	win.lock.Lock()
	defer win.lock.Unlock()
	texture, err := newTexture(win.Img)
	if err != nil {
		panic(err)
	}

	// Render it
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	// Display it
	window.SwapBuffers()
}

var program, vao uint32

func glInitialize() {
	// Initialize OpenGL bindings (must be done in an active OpenGL context)
	// and shaders
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Compile the vertex and fragment shaders into a program
	var err error
	program, err = newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)
	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Configure the vertex data - X, Y, Z, U, V
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(squareVertices)*4, gl.Ptr(squareVertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	// Tidy up
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newTexture(rgba *image.RGBA) (uint32, error) {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

var vertexShader = `
#version 330
in vec3 vert;
in vec2 vertTexCoord;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330
uniform sampler2D tex;
in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    outputColor = texture(tex, fragTexCoord);
}
` + "\x00"

// Square filling [-1,1]^2 as a triangle strip
// Image flipped in Y
var squareVertices = []float32{
	//  X, Y, Z, U, V
	-1, 1, 0, 0, 0,
	-1, -1, 0, 0, 1,
	1, 1, 0, 1, 0,
	1, -1, 0, 1, 1,
}

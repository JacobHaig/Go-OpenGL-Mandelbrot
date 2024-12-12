package main

import (
	"log"
	"os"
	"runtime"
	"time"

	gl "github.com/go-gl/gl/v4.1-core/gl"
	glfw "github.com/go-gl/glfw/v3.3/glfw"

	"fmt"
	"strings"
)

const (
	width       float32 = 1920
	height      float32 = 1080
	fps         int     = 1
	windowTitle         = "Mandelbrot"
)

var (
	// Is it possible to statically build these into the EXE with out external tools?
	fragmentShaderSource, _ = os.ReadFile("shaders/fragmentShaderSource.glsl")
	vertexShaderSource, _   = os.ReadFile("shaders/vertexShaderSource.glsl")

	triangle = []float32{ // Currenty unused
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
	}

	square = []float32{
		-1.0, 1.0, 0,
		-1.0, -1.0, 0,
		1.0, -1.0, 0,

		-1.0, 1.0, 0,
		1.0, 1.0, 0,
		1.0, -1.0, 0,
	}
)

func main() {
	// OpenGL needs to call the same operating system thread
	// more than once, there fore we lock it
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	// Vao is the Vertex array object, which is a ID that points to
	// a Vertex Buffer on the GPU that contains the Vertexs of the Square
	vertexArrayObject := makeVao(square)

	for !window.ShouldClose() {
		t := time.Now()

		// Draw Square on screen
		draw(vertexArrayObject, window, program)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}

	gl.DeleteProgram(program)
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(int(width), int(height), windowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	// Shader creation
	vertexShader, err := compileShader(string(vertexShaderSource)+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(string(fragmentShaderSource)+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Shader Clean up
	gl.ValidateProgram(program)
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

// Draw, draws the specified Vertex Array Object to the window.
// In this call, we are also checking for window interactions
// including mouse movement and keyboard input, but are throwing it away.
func draw(vertexArrayObject uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	// Uniforms must be set before every draw
	cstring, drop := gl.Strs("ress")
	ResolutionLoc := gl.GetUniformLocation(program, *cstring)
	if ResolutionLoc == -1 {
		println("ResolutionLoc - Error locating Uniform -", ResolutionLoc)
	}
	gl.Uniform2f(ResolutionLoc, width, height)
	drop()

	gl.BindVertexArray(vertexArrayObject)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}

// makeVao initializes and returns a vertex array from the points provided.
// makeVao returns the ID that points to the Vertex Buffer on the GPU
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

// Compiles the shader before it can be used. The shader ID gets
// returned if the shader compiles correctly.
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

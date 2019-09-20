#version 410

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

// This vs doesnt use all of these but this helps to keep it more generic
// and allow more code reuse
layout(location = 0) in vec3 vertex;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 uv;
layout(location = 3) in vec4 tangent;
layout(location = 4) in vec4 inColor;

out vec4 color;

void main(void)
{
    vec4 pp = projection * view * model * vec4(vertex, 1.0);
    color = inColor;
    gl_Position = pp;
};
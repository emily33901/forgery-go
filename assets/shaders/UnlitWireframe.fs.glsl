#version 410
#extension GL_OES_standard_derivatives : enable

uniform float lineWidth;
// uniform vec4 color;
uniform float blendFactor; //1.5..2.5

in vec3 coord;
in vec3 bary;
in vec4 color;

out vec4 fragCol;

void main () {
    // Pick a coordinate to visualize in a grid
    // vec3 scale = mod(extents, 50);
    vec3 scaled = vec3(bary);

    // Compute anti-aliased world-space grid lines
    // vec3 grid = abs(fract(coord - 0.5) - 0.5) / fwidth(coord);
    // float scale = extents;
    // vec3 arg = mod(scaled - 0.5*scale, scale);
    vec3 grid = abs(fract(scaled - 0.5) - 0.5) / fwidth(scaled);
    float line = min(min(grid.x, grid.y), grid.z);
    // float line = min(min(coord.x, coord.y), coord.z);

    // fragCol = color
    // fragCol = color;

    if (line > lineWidth) {
        discard;
    }
    fragCol = color;
}
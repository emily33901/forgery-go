#version 150

layout(triangles) in;
// layout(triangle_strip, max_vertices=3) out;
layout(line_strip, max_vertices=4) out;

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

out vec3 coord;
out vec3 bary;
out vec4 color;

in VertData {
    vec3 coord;
    vec4 color;
} inData[];

void main()
{
    vec3 size = vec3(0);

    vec3 p0 = inData[0].coord;
    vec3 p1 = inData[1].coord;
    vec3 p2 = inData[2].coord;

    float edgeA = length(p0 - p1);
    float edgeB = length(p1 - p2);
    float edgeC = length(p2 - p0);

    vec3 bary0 = vec3(0);
    vec3 bary1 = vec3(0);
    vec3 bary2 = vec3(0);

    int out0;
    int out1;
    int out2;

    if (edgeA > edgeB && edgeA > edgeC) {
        // edgeA is hyp
        out0 = 0;
        out1 = 2;
        out2 = 1;
    } else if (edgeB > edgeC && edgeB > edgeA) {
        // edgeB is hyp
        out0 = 1;
        out1 = 0;
        out2 = 2;
    } else {
        // edgeC is hyp
        out0 = 2;
        out1 = 1;
        out2 = 0;
    }

    // Now emit vertices
    coord = inData[out0].coord;
    color = inData[out0].color;
    gl_Position = gl_in[out0].gl_Position;
    EmitVertex();

    coord = inData[out1].coord;
    color = inData[out1].color;
    gl_Position = gl_in[out1].gl_Position;
    EmitVertex();

    coord = inData[out1].coord;
    color = inData[out1].color;
    gl_Position = gl_in[out1].gl_Position;
    EmitVertex();

    coord = inData[out2].coord;
    color = inData[out2].color;
    gl_Position = gl_in[out2].gl_Position;
    EmitVertex();

#if 0
    // This is the magic for doing this on triangles
    // or arbitrary width lines
    // @TODO revisit this

    // We need to keep all the edges numbered "consistently"
    // By making the edge away from the hyp 

    if (edgeA > edgeB && edgeA > edgeC) {
        // edgeA is hyp
        bary0 = vec3(1.0, 0.0, 1.0);
        bary1 = vec3(0.0, 1.0, 0.0);
        bary2 = vec3(1.0, 0.0, 0.0);
    } else if (edgeB > edgeC && edgeB > edgeA) {
        // edgeB is hyp
        bary0 = vec3(0.0, 1.0, 0.0);
        bary1 = vec3(1.0, 0.0, 1.0);
        bary2 = vec3(0.0, 0.0, 1.0);
    } else {
        // edgeC is hyp
        bary0 = vec3(0.0, 0.0, 1.0);
        //bary1 = vec3(1.0, 0.0, 0.0);
        bary2 = vec3(1.0, 0.0, 0.0);
    }

    extents = size;

    // Actually emit verticies

    gl_Position = gl_in[0].gl_Position;
    coord = p0.xyz;
    bary = bary0;
    EmitVertex();

    gl_Position = gl_in[1].gl_Position;
    coord = p1.xyz;
    bary = bary1;
    EmitVertex();

    gl_Position = gl_in[2].gl_Position;
    coord = p2.xyz;
    bary = bary2;
    EmitVertex();

#endif
}
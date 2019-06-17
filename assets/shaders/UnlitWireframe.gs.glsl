#version 150

layout(triangles) in;
layout(triangle_strip, max_vertices=3) out;

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

out vec3 coord;
out vec3 extents;
out vec3 bary;

in VertData {
    vec3 coord;
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
}
#version 410

uniform sampler2D albedoSampler;

uniform bool shouldDiscard;

in vec2 UV;
in vec3 dist;

out vec4 frag_colour;

void AddAlbedo(inout vec4 fragColour, in sampler2D sampler, in vec2 uv) 
{
    fragColour = texture(sampler, uv).rgba;
}

void main() {
    if (shouldDiscard == false) {
        AddAlbedo(frag_colour, albedoSampler, UV);

        //discard;
        //frag_colour = gl_Color;
        //frag_colour.a = 0;
    }
}
#version 410

out vec4 frag_color;
in vec2 fragCoord;

uniform vec2 res;

float normalize(float value, float from_min, float from_max, float to_min, float to_max) {
    return (to_min + ((value- from_min) * (to_max - to_min)) / (from_max - from_min));
}

float iterateMandelbrot2(float c_re, float c_im, int max_iter){
    float z_re = 0.0;
    float z_im = 0.0;

    int totalIter = 0;
    for( int i = 0; i < max_iter; i ++) {
        if (z_re * z_re + z_im * z_im > 4.0) {  break;   }

        float new_re = (z_re * z_re) - (z_im * z_im);
        float new_im = (z_re * z_im) + (z_im * z_re);

        z_re = new_re + c_re;
        z_im = new_im + c_im;

        totalIter += 1;
    }
    return totalIter;
}


void main() {
    // frag_color is of R, G, B, A type. Each value ranges between 0..1
    float x = normalize( gl_FragCoord.x, 0, res.x, -2.0,  1.0);
    float y = normalize( gl_FragCoord.y, 0, res.y,  1.0, -1.0);

    float totalIter = iterateMandelbrot2(x, y, 25) / 25.0;

    frag_color = vec4(totalIter, totalIter, totalIter , 1); 
}



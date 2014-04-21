#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <fcntl.h>
#include <string>
#include <iostream>

#include "srl_decoder.h"
#include <dlfcn.h>

PerlInterpreter* my_perl;
typedef srl_decoder_t* build_decoder_struct(pTHX_ HV *opt);
typedef SV* decode_into(pTHX_ srl_decoder_t *dec, SV *src, SV *body_into, UV start_offset);

int main(int argc, char** argv) {
    std::string content;
    struct stat st;

    int fd = open(argv[1], O_RDONLY);
    fstat(fd, &st);

    char buf[4096];
    while (st.st_size > 0) {
        ssize_t rlen = read(fd, buf, sizeof(buf));
        content.append(buf, rlen);
        st.st_size -= rlen;
    }

    my_perl = perl_alloc();
    perl_construct(my_perl);

    SV* src = newSVpvn(content.data(), content.size());
    SV* into = newSV(0);
    std::cout << src->sv_refcnt << std::endl;
    std::cout << src->sv_u.svu_pv << std::endl;

    void* h = dlopen("/root/Sereal/Perl/Decoder/libdecoder.so", RTLD_NOW);
    if (h == NULL) {
        std::cout << dlerror() << std::endl;
    }

    void* build_decoder = dlsym(h, "srl_build_decoder_struct");
    if (build_decoder == NULL) std::cout << dlerror() << std::endl;
    std::cout << build_decoder << std::endl;

    void* dec_into = dlsym(h, "srl_decode_into");
    if (dec_into == NULL) std::cout << dlerror() << std::endl;
    std::cout << dec_into << std::endl;

    srl_decoder_t* dec = (* (build_decoder_struct*) build_decoder)(my_perl, 0);
    std::cout << dec << std::endl;
    std::cout << dec->max_recursion_depth << std::endl;

    for (int i = 0; i < 5; ++i) {
        UV offset = 0;
        SV* res = (* (decode_into*) dec_into)(my_perl, dec, src, NULL, offset);

        std::cout << res << std::endl;
        std::cout << res->sv_refcnt << std::endl;
        std::cout << SvTYPE(SvRV(res)) << std::endl;

        sv_free(res);
        //AV* array = (AV*) SvRV(res);
        //std::cout << array->sv_any->xav_fill << std::endl;
    }

    //SV** array = res->sv_u.svu_array;
}


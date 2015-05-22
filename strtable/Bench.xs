#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "ppport.h"

#include "strtable.h"

MODULE = Bench        PACKAGE = Bench

void
benchmark_strtable(data, size)
    AV *data;
    UV size;
  PREINIT:
    SV *sv;
    STRLEN len;
    const char *str;
    SSize_t i, avlen;
    strtable_t tbl;
  CODE:
    avlen = av_len(data) + 1;
    strtable_init(&tbl, (uint32_t) size);

    for (i = 0; i < avlen; ++i) {
        sv = *av_fetch(data, i, 0);
        str = SvPV(sv, len);
        strtable_insert(&tbl, str, (uint32_t) len, (uint32_t) i);
    }

    if (tbl.keys != avlen)
        warn("benchmark_strtable: tbl->key != avlen (%d != %d)", (int) tbl.keys, (int) avlen);

    strtable_clear(&tbl);

void 
benchmark_hv(data)
    AV *data;
    UV size;
  PREINIT:
    SV *sv;
    HV *hv;
    STRLEN len;
    const char *str;
    SSize_t i, avlen;
  CODE:
    hv = newHV();
    avlen = av_len(data) + 1;

    for (i = 0; i < avlen; ++i) {
        sv = *av_fetch(data, i, 0);
        str = SvPV(sv, len);
        hv_fetch(hv, str, len, 1);
    }

    if (av_len((AV*) hv) != avlen)
        warn("benchmark_hv: keys != avlen (%d != %d)", (int) av_len((AV*) hv), (int) avlen);

    hv_undef(hv);

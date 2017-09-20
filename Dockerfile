FROM platina-buildroot

RUN rm -rf builds/*/build/platina-goes-master/build && rm -f builds/*/build/platina-goes-master/.stamp_built builds/*/build/platina-goes-master/.stamp_target_installed builds/*/build/platina-goes-master/.stamp_staging_installed

COPY . builds/example-amd64/build/platina-goes-master
COPY . builds/platina-mk1/build/platina-goes-master
COPY . builds/platina-mk1-bmc/build/platina-goes-master

RUN make -C buildroot O=../builds/example-amd64 example-amd64_defconfig && make -C builds/example-amd64 all
RUN make -C buildroot O=../builds/platina-mk1 platina-mk1_defconfig && make -C builds/platina-mk1 all
RUN make -C buildroot O=../builds/platina-mk1-bmc platina-mk1-bmc_defconfig && make -C builds/platina-mk1-bmc all

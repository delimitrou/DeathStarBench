perf stat -C 1 -e uops_retired.slots:u,uops_retired.slots:k,instructions:u,instructions:k -- sleep 60 > uops-ampere.log 2>&1
perf stat -C 1 -M DSB_Coverage -- sleep 60 > dsb-ampere.log 2>&1
/var/services/homes/shanqing-epfl/pmu-tools/toplev -l4 --core C1 -v --no-desc --no-multiplex --xlsx topdown-ampere.xlsx -- sleep 30
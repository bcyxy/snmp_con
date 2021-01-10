go build

rm -rf output/
mkdir -p output/
mv testvm output/
cp conf.json output/
touch output/dev_list.txt

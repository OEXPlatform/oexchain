# 脚本用法：
# ./getFNode.sh 1 10 : 获取矿工1~矿工10的FNode信息

oldifs="$IFS"
IFS=$'\n'

function getFNode ()
{
    fnode=""
    logFile="logs/minernodetest"$1".log"
    for line in `head -50 $logFile`
    do
            len=${#line}
            if [[ $len -lt 100 ]]; then
                    continue
            fi
            #line=`expr substr $line 50 200`
            tmpFNode=`expr $line : '.*self=\(.*\)'`
            fnodeLen=${#tmpFNode}
	    if [[ $fnodeLen -gt 10 ]]; then
		echo minernodetest$1:$tmpFNode
		break
	    fi
    done
}

if [[ $# -ne 2 ]]; then
        echo "command=>$0, no parameters"
        exit 1
fi

startNodeNum=$1
while(( $startNodeNum<=$2 ))
do
        getFNode $startNodeNum
        let "startNodeNum++"
done


IFS="$oldifs"

package c3

const Index = `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title> c3-demo </title>
        <!-- Load c3.css -->
        <style type="text/css">
          {{.C3CSS}}
        </style>

        <!-- Load d3.js and c3.js -->
        <script >
         {{.JqueryJS}}
        </script>
        <script charset="utf-8">
         {{.D3JS}}
        </script>
        <script >
          {{.C3JS}}
        </script>
    </head>
    <body>
        <div id="chart"></div>
        <script >
            $(function(){
                var chart = c3.generate({
                    bindto: '#chart',
                    size: {
                        height: 640,
                        width: 1350,
                    },
                    data: {
                        columns: [
                            ['并发数', {{.ConcurrentStr}}],
                            ['响应时间', {{.ResponseTimeStr}}],
                            ['Pod 数', {{.PodNumStr}}]
                        ],
                        type: 'line',
                        axes: {
                            '并发数': 'y',
                            'Pod 数': 'y',
                            '响应时间': 'y2'
                        }
                    },
                    axis: {
                        y: {
                            padding: {top: 200, bottom: 0}
                        },
                        y2: {
                            padding: {top: 100, bottom: 0},
                            show: true
                        }
                    },
                    //subchart: {
                    //    show: true
                    //}
                });
            });
        </script>

    </body>
</html>
`

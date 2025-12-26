import { DATE_TIME_FORMAT, DateSelectOption } from '@/constants/datetime';
import { listSqlMetrics, querySqlDetailInfo } from '@/services/sql';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { ProCard, ProDescriptions, ProTable } from '@ant-design/pro-components';
import { Line } from '@antv/g2plot';
import { history, useParams, useRequest, useSearchParams } from '@umijs/max';
import { Button, DatePicker, Select, Space, Tag, Typography } from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import dayjs from 'dayjs';
import { useEffect, useRef, useState } from 'react';
import { getLocale } from 'umi';

const { RangePicker } = DatePicker;
const { Title } = Typography;

// --- Chart Component ---

interface SqlTrendChartProps {
  data: API.MetricData[];
  type: 'execution' | 'latency';
  height?: number;
}

const SqlTrendChart: React.FC<SqlTrendChartProps> = ({
  data,
  type,
  height = 300,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<Line>();

  useEffect(() => {
    if (!containerRef.current || !data || data.length === 0) return;

    // Transform data for G2Plot
    // API.MetricData: { metric: { name, labels }, values: { timestamp, value }[] }
    // G2Plot expects array of objects
    const plotData: any[] = [];
    data.forEach((metricData) => {
      const name = metricData.metric?.name || type;
      metricData.values?.forEach((v) => {
        plotData.push({
          time: v.timestamp * 1000, // Convert to ms
          value: v.value,
          category: name,
        });
      });
    });

    // sort by time
    plotData.sort((a, b) => a.time - b.time);

    if (chartRef.current) {
      chartRef.current.destroy();
    }

    const chart = new Line(containerRef.current, {
      data: plotData,
      xField: 'time',
      yField: 'value',
      seriesField: 'category',
      height,
      xAxis: {
        type: 'time',
        mask: 'HH:mm',
      },
      yAxis: {
        label: {
          formatter: (v: string) => {
            return Number(v).toFixed(2);
          },
        },
      },
      tooltip: {
        formatter: (datum: any) => {
          return { name: datum.category, value: datum.value.toFixed(2) };
        },
      },
      legend: {
        position: 'top',
      },
    });

    chart.render();
    chartRef.current = chart;

    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
      }
    };
  }, [data, type, height]);

  return <div ref={containerRef} />;
};

// --- Main Page Component ---

const SqlDetail: React.FC = () => {
  const { ns, name, tenantName, sqlId } = useParams<{
    ns: string;
    name: string;
    tenantName: string;
    sqlId: string;
  }>();
  const [searchParams] = useSearchParams();
  const dbName = searchParams.get('dbName') || '';
  const urlStartTime = searchParams.get('startTime');
  const urlEndTime = searchParams.get('endTime');

  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    urlStartTime
      ? dayjs.unix(Number(urlStartTime))
      : dayjs().subtract(30, 'minute'),
    urlEndTime ? dayjs.unix(Number(urlEndTime)) : dayjs(),
  ]);

  const [latencyMetricsMeta, setLatencyMetricsMeta] = useState<
    API.SqlMetricMeta[]
  >([]);
  const [selectedLatencyMetrics, setSelectedLatencyMetrics] = useState<
    string[]
  >([]);

  // Fetch metrics meta to populate latency selector
  useRequest(
    () =>
      listSqlMetrics({ language: getLocale() === 'zh-CN' ? 'zh_CN' : 'en_US' }),
    {
      onSuccess: (data) => {
        const list = Array.isArray(data) ? data : (data as any)?.data || [];
        const latencyCat = list.find((c: any) => c.category === 'latency');
        if (latencyCat && latencyCat.metrics) {
          setLatencyMetricsMeta(latencyCat.metrics);
          // Set default selected metrics based on displayByDefault
          const defaults = latencyCat.metrics
            .filter((m: any) => m.displayByDefault)
            .map((m: any) => m.key);
          if (defaults.length > 0) {
            setSelectedLatencyMetrics(defaults);
          } else {
            // Fallback to first if no defaults
            if (latencyCat.metrics.length > 0) {
              setSelectedLatencyMetrics([latencyCat.metrics[0].key]);
            }
          }
        }
      },
    },
  );

  const { data: detailData, loading } = useRequest(
    async () => {
      if (!ns || !name || !sqlId) return;
      const start = timeRange[0].unix();
      const end = timeRange[1].unix();
      // Calculate interval: aim for ~60 points? or just 60s default?
      // Use (end - start) / 60, min 1s
      let interval = Math.floor((end - start) / 60);
      if (interval < 1) interval = 1;

      return querySqlDetailInfo({
        namespace: ns,
        obtenant: name,
        sqlId,
        database: dbName,
        startTime: start,
        endTime: end,
        interval,
        outputColumns: selectedLatencyMetrics,
      });
    },
    {
      refreshDeps: [timeRange, selectedLatencyMetrics, ns, name, sqlId, dbName],
    },
  );

  // Time Range Helpers
  const range = (start: number, end: number) => {
    const result = [];
    for (let i = start; i < end; i++) {
      result.push(i);
    }
    return result;
  };

  const disabledDateTime: RangePickerProps['disabledTime'] = (_) => {
    const isToday = _?.date() === dayjs().date();
    if (!isToday)
      return {
        disabledHours: () => [],
        disabledMinutes: () => [],
        disabledSeconds: () => [],
      };
    return {
      disabledHours: () => range(0, 24).splice(dayjs().hour() + 1, 24),
      disabledMinutes: (hour) => {
        if (hour === dayjs().hour()) {
          return range(0, 60).splice(dayjs().minute() + 1, 60);
        }
        return [];
      },
      disabledSeconds: (hour, minute) => {
        if (hour === dayjs().hour() && minute === dayjs().minute()) {
          return range(0, 60).splice(dayjs().second(), 60);
        }
        return [];
      },
    };
  };

  const disabledDate: RangePickerProps['disabledDate'] = (current) => {
    return current && current > dayjs().endOf('day');
  };

  const handleTimeChange = (dates: any) => {
    if (dates) {
      setTimeRange([dates[0], dates[1]]);
    }
  };

  const { data: sqlMetaData } = useRequest(
    async () => {
      if (!ns || !name || !sqlId) return;
      return listSqlStats({
        // Reuse existing list API to get meta
        namespace: ns,
        obtenant: name,
        startTime: timeRange[0].unix(),
        endTime: timeRange[1].unix(),
        keyword: sqlId, // Filter by SQL ID
        outputColumns: ['query_sql', 'user_name', 'db_name', 'sql_id'], // metrics
        pageSize: 1,
        pageNum: 1,
      } as any);
    },
    {
      refreshDeps: [ns, name, sqlId, timeRange],
    },
  );

  const sqlMeta = sqlMetaData?.data?.items?.[0];

  return (
    <div
      style={{
        backgroundColor: '#F5F7FA',
        minHeight: '100vh',
        padding: 24,
      }}
    >
      {/* Header & Basic Info */}
      <ProCard ghost gutter={[16, 16]} direction="column">
        <ProCard loading={loading}>
          <Space
            style={{
              marginBottom: 16,
              justifyContent: 'space-between',
              width: '100%',
            }}
          >
            <Space>
              <Button
                icon={<ArrowLeftOutlined />}
                onClick={() => history.back()}
              >
                Back
              </Button>
              <Title level={4} style={{ margin: 0 }}>
                SQL Detail: {sqlId}
              </Title>
            </Space>
            <RangePicker
              showTime
              value={timeRange}
              onChange={handleTimeChange}
              format={DATE_TIME_FORMAT}
              disabledDate={disabledDate}
              disabledTime={disabledDateTime}
              presets={DateSelectOption.filter((o) => o.value !== 'custom').map(
                (o) => ({
                  label: o.label,
                  value: [dayjs().subtract(o.value as number, 'ms'), dayjs()],
                }),
              )}
            />
          </Space>

          <ProDescriptions column={3} bordered>
            <ProDescriptions.Item label="SQL Text" span={3}>
              <Typography.Paragraph
                ellipsis={{ rows: 2, expandable: true, symbol: 'more' }}
                copyable
              >
                {sqlMeta?.querySql || '-'}
              </Typography.Paragraph>
            </ProDescriptions.Item>
            <ProDescriptions.Item label="SQL ID">{sqlId}</ProDescriptions.Item>
            <ProDescriptions.Item label="Database">
              {dbName || sqlMeta?.dbName || '-'}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="User">
              {sqlMeta?.userName || '-'}
            </ProDescriptions.Item>
          </ProDescriptions>
        </ProCard>

        {/* Charts */}
        <ProCard title="History Request Info" headerBordered loading={loading}>
          <ProCard split="vertical">
            <ProCard title="Executions" colSpan={12}>
              <SqlTrendChart
                data={detailData?.data?.executionTrend || []}
                type="execution"
              />
            </ProCard>
            <ProCard
              title={
                <Space>
                  <span>Latency</span>
                  <Select
                    mode="multiple"
                    maxTagCount="responsive"
                    value={selectedLatencyMetrics}
                    onChange={setSelectedLatencyMetrics}
                    style={{ width: 400 }}
                    options={latencyMetricsMeta.map((m) => ({
                      label: m.name,
                      value: m.key,
                    }))}
                  />
                </Space>
              }
              colSpan={12}
            >
              <SqlTrendChart
                data={detailData?.data?.latencyTrend || []}
                type="latency"
              />
            </ProCard>
          </ProCard>
        </ProCard>

        {/* Diagnosis */}
        <ProCard title="Diagnosis & Advice" headerBordered loading={loading}>
          {detailData?.data?.diagnoseInfo &&
          detailData.data.diagnoseInfo.length > 0 ? (
            detailData.data.diagnoseInfo.map((diag, idx) => (
              <Alert
                key={idx}
                message={diag.reason}
                description={diag.suggestion}
                type="warning"
                showIcon
                style={{ marginBottom: 8 }}
              />
            ))
          ) : (
            <Tag color="green">No diagnosis issues found.</Tag>
          )}
        </ProCard>

        {/* Plan Statistics */}
        <ProCard title="Plan Statistics" headerBordered loading={loading}>
          <ProTable<API.PlanStatistic>
            dataSource={detailData?.data?.plans || []}
            columns={[
              {
                title: 'Plan ID',
                dataIndex: 'planID',
                render: (text, record) => (
                  <a
                    href={`/tenant/${ns}/${name}/${tenantName}/sql/${sqlId}/plan/${record.planID}`}
                  >
                    {text}
                  </a>
                ),
              },
              { title: 'Svr IP', dataIndex: 'svrIP' },
              { title: 'Svr Port', dataIndex: 'svrPort' },
              { title: 'Plan Hash', dataIndex: 'planHash' },
              { title: 'Cost', dataIndex: 'cost' },
              { title: 'CPU Cost', dataIndex: 'cpuCost' },
              { title: 'IO Cost', dataIndex: 'ioCost' },
              {
                title: 'Generated Time',
                dataIndex: 'generatedTime',
                valueType: 'dateTime',
              },
            ]}
            search={false}
            toolBarRender={false}
            pagination={{ pageSize: 5 }}
          />
        </ProCard>

        {/* Index Info */}
        <ProCard title="Index Info" headerBordered loading={loading}>
          <ProTable<API.IndexInfo>
            dataSource={detailData?.data?.indexies || []}
            columns={[
              { title: 'Table Name', dataIndex: 'tableName' },
              { title: 'Index Name', dataIndex: 'indexName' },
              { title: 'Index Type', dataIndex: 'indexType' },
              { title: 'Uniqueness', dataIndex: 'uniqueness' },
              {
                title: 'Columns',
                dataIndex: 'columns',
                render: (cols) =>
                  Array.isArray(cols) ? cols.join(', ') : cols,
              },
              { title: 'Status', dataIndex: 'status' },
            ]}
            search={false}
            toolBarRender={false}
            pagination={{ pageSize: 5 }}
          />
        </ProCard>
      </ProCard>
    </div>
  );
};

export default SqlDetail;

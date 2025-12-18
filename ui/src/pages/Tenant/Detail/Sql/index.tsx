import { DATE_TIME_FORMAT, DateSelectOption } from '@/constants/datetime';
import { listSqlMetrics, listSqlStats } from '@/services/sql';
import { SettingOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { useParams, useRequest } from '@umijs/max';
import { Button } from 'antd';
import type { RangePickerProps } from 'antd/es/date-picker';
import dayjs from 'dayjs';
import { useMemo, useRef, useState } from 'react';
import { getLocale } from 'umi';
import ColumnSelectionDrawer from './ColumnSelectionDrawer';

export default function SqlList() {
  const { ns, name, tenantName } = useParams<{
    ns: string;
    name: string;
    tenantName: string;
  }>();
  const actionRef = useRef<ActionType>();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedMetricKeys, setSelectedMetricKeys] = useState<string[]>([]);

  // Helper to robustly extract metrics array regardless of response format
  const getMetricsList = (data: any): API.SqlMetricMetaCategory[] => {
    if (!data) return [];
    if (Array.isArray(data)) return data;
    if (data.data && Array.isArray(data.data)) return data.data;
    return [];
  };

  // Fetch metric metadata to know available columns and defaults
  const { data: metricsData } = useRequest(
    () =>
      listSqlMetrics({ language: getLocale() === 'zh-CN' ? 'zh_CN' : 'en_US' }),
    {
      onSuccess: (data) => {
        const list = getMetricsList(data);
        const defaults: string[] = [];
        list.forEach((category) => {
          category.metrics.forEach((metric) => {
            if (metric.displayByDefault || metric.immutable) {
              defaults.push(metric.key);
            }
          });
        });
        setSelectedMetricKeys(defaults);
      },
    },
  );

  const initialTimeRange: [dayjs.Dayjs, dayjs.Dayjs] = [
    dayjs().subtract(30, 'minute'),
    dayjs(),
  ];

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

  const metaFieldMap: Record<string, keyof API.SqlMetaInfo> = {
    query_sql: 'querySql',
    db_name: 'dbName',
    user_name: 'userName',
    sql_id: 'sqlId',
    svr_ip: 'svrIp',
    svr_port: 'svrPort',
    client_ip: 'clientIp',
  };

  // Generate dynamic columns based on selected keys and metadata
  const dynamicColumns: ProColumns<API.SqlInfo>[] = useMemo(() => {
    const list = getMetricsList(metricsData);
    if (list.length === 0 || selectedMetricKeys.length === 0) return [];

    const cols: ProColumns<API.SqlInfo>[] = [];
    const allMetrics: API.SqlMetricMeta[] = [];
    list.forEach((cat) => {
      allMetrics.push(...cat.metrics);
    });

    allMetrics.forEach((metric) => {
      if (selectedMetricKeys.includes(metric.key)) {
        const colConfig: ProColumns<API.SqlInfo> = {
          title: metric.name,
          dataIndex: metric.key,
          search: false,
          width: 120,
        };

        if (metaFieldMap[metric.key]) {
          colConfig.dataIndex = metaFieldMap[metric.key];
          if (metric.key === 'sql_id') {
            colConfig.fixed = 'left';
            colConfig.width = 150;
            colConfig.copyable = true;
            colConfig.ellipsis = true;
            colConfig.render = (dom, record) => (
              <a
                href={`/tenant/${ns}/${name}/${tenantName}/sql/${record.sqlId}?dbName=${record.dbName}`}
              >
                {dom}
              </a>
            );
          } else if (metric.key === 'query_sql') {
            colConfig.width = 200;
            colConfig.ellipsis = true;
          } else if (metric.key === 'user_name') {
            colConfig.width = 100;
          }
        } else {
          colConfig.render = (_, record) => {
            const stat =
              record.executionStatistics?.find((s) => s.name === metric.key) ||
              record.latencyStatistics?.find((s) => s.name === metric.key);
            return stat ? stat.value : '-';
          };
        }
        cols.push(colConfig);
      }
    });
    return cols;
  }, [metricsData, selectedMetricKeys, ns, name, tenantName]);

  const columns: ProColumns<API.SqlInfo>[] = [
    ...dynamicColumns,
    {
      title: 'Time Range',
      dataIndex: 'timeRange',
      valueType: 'dateTimeRange',
      hideInTable: true,
      order: 1, // Lowest priority -> last in search form
      fieldProps: {
        format: DATE_TIME_FORMAT,
        disabledDate: disabledDate,
        disabledTime: disabledDateTime,
        presets: DateSelectOption.filter((o) => o.value !== 'custom').map(
          (o) => ({
            label: o.label,
            value: [dayjs().subtract(o.value as number, 'ms'), dayjs()],
          }),
        ),
      },
      search: {
        transform: (value: [string, string]) => {
          return {
            startTime: dayjs(value[0]).unix(),
            endTime: dayjs(value[1]).unix(),
          };
        },
      },
    },
  ];

  return (
    <>
      <ProTable<API.SqlInfo>
        headerTitle="SQL Analysis"
        actionRef={actionRef}
        rowKey="sqlId"
        form={{
          initialValues: {
            timeRange: initialTimeRange,
          },
        }}
        search={{
          collapsed: false,
          collapseRender: false,
          labelWidth: 'auto',
        }}
        toolBarRender={() => [
          <Button
            key="column-selection"
            icon={<SettingOutlined />}
            onClick={() => setDrawerOpen(true)}
          >
            Column Selection
          </Button>,
        ]}
        scroll={{ x: 'max-content' }}
        request={async (params, sort) => {
          if (!ns || !name || !tenantName) {
            return { data: [], success: false };
          }

          const { startTime, endTime, ...restParams } = params;

          // Ensure startTime and endTime are present, defaulting to initialTimeRange if not
          const effectiveStartTime = startTime ?? initialTimeRange[0].unix();
          const effectiveEndTime = endTime ?? initialTimeRange[1].unix();

          const msg = await listSqlStats({
            namespace: ns,
            obtenant: name,
            sortColumn: Object.keys(sort)[0],
            sortOrder: Object.values(sort)[0] === 'ascend' ? 'asc' : 'desc',
            pageNum: restParams.current,
            pageSize: restParams.pageSize,
            keyword: restParams.querySql as string,
            startTime: effectiveStartTime,
            endTime: effectiveEndTime,
            outputColumns: selectedMetricKeys,
          });

          return {
            data: msg.data,
            success: msg.successful,
            total: msg.data?.length || 0,
          };
        }}
        columns={columns}
        pagination={{
          pageSize: 10,
        }}
      />
      <ColumnSelectionDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        selectedKeys={selectedMetricKeys}
        onSelectionChange={(keys) => {
          setSelectedMetricKeys(keys);
          actionRef.current?.reload();
        }}
        metrics={getMetricsList(metricsData)}
      />
    </>
  );
}

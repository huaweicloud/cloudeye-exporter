package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// ======================== Label Tests ========================

func TestLabel_GetNameGetValue(t *testing.T) {
	label := &Label{Name: "instance_id", Value: "abc-123"}
	assert.Equal(t, "instance_id", label.GetName())
	assert.Equal(t, "abc-123", label.GetValue())
}

func TestLabel_NilReceiver(t *testing.T) {
	var label *Label
	assert.Equal(t, "", label.GetName())
	assert.Equal(t, "", label.GetValue())
}

func TestLabel_EmptyFields(t *testing.T) {
	label := &Label{}
	assert.Equal(t, "", label.GetName())
	assert.Equal(t, "", label.GetValue())
}

func TestLabel_Reset(t *testing.T) {
	label := &Label{Name: "host", Value: "server01"}
	label.Reset()
	assert.Equal(t, "", label.Name)
	assert.Equal(t, "", label.Value)
}

func TestLabel_String(t *testing.T) {
	label := &Label{Name: "region", Value: "cn-north-1"}
	s := label.String()
	assert.NotEmpty(t, s)
}

func TestLabel_ProtoReflect(t *testing.T) {
	label := &Label{Name: "az", Value: "az1"}
	msg := label.ProtoReflect()
	assert.NotNil(t, msg)
	assert.Equal(t, "name", string(msg.Descriptor().Fields().ByName("name").Name()))
}

func TestLabel_ProtoReflect_NilReceiver(t *testing.T) {
	var label *Label
	msg := label.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestLabel_Descriptor(t *testing.T) {
	desc, path := (*Label)(nil).Descriptor()
	assert.NotNil(t, desc)
	assert.Equal(t, []int{2}, path)
}

func TestLabel_ProtoMessage(t *testing.T) {
	// 验证 Label 实现了 proto.Message 接口
	var _ interface{ ProtoMessage() } = &Label{}
}

// ======================== Sample Tests ========================

func TestSample_GetValueGetTimestamp(t *testing.T) {
	sample := &Sample{Value: 99.5, Timestamp: 1700000000000}
	assert.Equal(t, 99.5, sample.GetValue())
	assert.Equal(t, int64(1700000000000), sample.GetTimestamp())
}

func TestSample_NilReceiver(t *testing.T) {
	var sample *Sample
	assert.Equal(t, float64(0), sample.GetValue())
	assert.Equal(t, int64(0), sample.GetTimestamp())
}

func TestSample_ZeroValues(t *testing.T) {
	sample := &Sample{}
	assert.Equal(t, float64(0), sample.GetValue())
	assert.Equal(t, int64(0), sample.GetTimestamp())
}

func TestSample_Reset(t *testing.T) {
	sample := &Sample{Value: 42.0, Timestamp: 1234567890}
	sample.Reset()
	assert.Equal(t, float64(0), sample.Value)
	assert.Equal(t, int64(0), sample.Timestamp)
}

func TestSample_String(t *testing.T) {
	sample := &Sample{Value: 1.5, Timestamp: 100}
	s := sample.String()
	assert.NotEmpty(t, s)
}

func TestSample_ProtoReflect(t *testing.T) {
	sample := &Sample{Value: 10.0, Timestamp: 200}
	msg := sample.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestSample_ProtoReflect_NilReceiver(t *testing.T) {
	var sample *Sample
	msg := sample.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestSample_Descriptor(t *testing.T) {
	desc, path := (*Sample)(nil).Descriptor()
	assert.NotNil(t, desc)
	assert.Equal(t, []int{3}, path)
}

func TestSample_NegativeValue(t *testing.T) {
	sample := &Sample{Value: -1.5, Timestamp: -100}
	assert.Equal(t, -1.5, sample.GetValue())
	assert.Equal(t, int64(-100), sample.GetTimestamp())
}

// ======================== TimeSeries Tests ========================

func TestTimeSeries_GetLabelsGetSamples(t *testing.T) {
	ts := &TimeSeries{
		Labels:  []*Label{{Name: "__name__", Value: "cpu_usage"}},
		Samples: []*Sample{{Value: 85.5, Timestamp: 1700000000000}},
	}
	assert.Equal(t, 1, len(ts.GetLabels()))
	assert.Equal(t, "__name__", ts.GetLabels()[0].Name)
	assert.Equal(t, 1, len(ts.GetSamples()))
	assert.Equal(t, 85.5, ts.GetSamples()[0].Value)
}

func TestTimeSeries_NilReceiver(t *testing.T) {
	var ts *TimeSeries
	assert.Nil(t, ts.GetLabels())
	assert.Nil(t, ts.GetSamples())
}

func TestTimeSeries_EmptyFields(t *testing.T) {
	ts := &TimeSeries{}
	assert.Nil(t, ts.GetLabels())
	assert.Nil(t, ts.GetSamples())
}

func TestTimeSeries_Reset(t *testing.T) {
	ts := &TimeSeries{
		Labels:  []*Label{{Name: "host", Value: "server01"}},
		Samples: []*Sample{{Value: 50.0, Timestamp: 100}},
	}
	ts.Reset()
	assert.Nil(t, ts.Labels)
	assert.Nil(t, ts.Samples)
}

func TestTimeSeries_String(t *testing.T) {
	ts := &TimeSeries{
		Labels:  []*Label{{Name: "job", Value: "ces"}},
		Samples: []*Sample{{Value: 1.0, Timestamp: 200}},
	}
	s := ts.String()
	assert.NotEmpty(t, s)
}

func TestTimeSeries_ProtoReflect(t *testing.T) {
	ts := &TimeSeries{}
	msg := ts.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestTimeSeries_ProtoReflect_NilReceiver(t *testing.T) {
	var ts *TimeSeries
	msg := ts.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestTimeSeries_Descriptor(t *testing.T) {
	desc, path := (*TimeSeries)(nil).Descriptor()
	assert.NotNil(t, desc)
	assert.Equal(t, []int{1}, path)
}

func TestTimeSeries_MultipleLabelsAndSamples(t *testing.T) {
	ts := &TimeSeries{
		Labels: []*Label{
			{Name: "__name__", Value: "cpu_usage"},
			{Name: "instance_id", Value: "i-001"},
			{Name: "az", Value: "az1"},
		},
		Samples: []*Sample{
			{Value: 10.0, Timestamp: 100},
			{Value: 20.0, Timestamp: 200},
		},
	}
	assert.Equal(t, 3, len(ts.GetLabels()))
	assert.Equal(t, 2, len(ts.GetSamples()))
}

// ======================== WriteRequest Tests ========================

func TestWriteRequest_GetTimeseries(t *testing.T) {
	wr := &WriteRequest{
		Timeseries: []*TimeSeries{
			{
				Labels:  []*Label{{Name: "__name__", Value: "mem_usage"}},
				Samples: []*Sample{{Value: 60.0, Timestamp: 1700000000000}},
			},
		},
	}
	assert.Equal(t, 1, len(wr.GetTimeseries()))
	assert.Equal(t, "mem_usage", wr.GetTimeseries()[0].Labels[0].Value)
}

func TestWriteRequest_NilReceiver(t *testing.T) {
	var wr *WriteRequest
	assert.Nil(t, wr.GetTimeseries())
}

func TestWriteRequest_EmptyTimeseries(t *testing.T) {
	wr := &WriteRequest{}
	assert.Nil(t, wr.GetTimeseries())
}

func TestWriteRequest_Reset(t *testing.T) {
	wr := &WriteRequest{
		Timeseries: []*TimeSeries{{Labels: []*Label{{Name: "a", Value: "b"}}}},
	}
	wr.Reset()
	assert.Nil(t, wr.Timeseries)
}

func TestWriteRequest_String(t *testing.T) {
	wr := &WriteRequest{
		Timeseries: []*TimeSeries{
			{
				Labels:  []*Label{{Name: "job", Value: "test"}},
				Samples: []*Sample{{Value: 1.0, Timestamp: 1}},
			},
		},
	}
	s := wr.String()
	assert.NotEmpty(t, s)
}

func TestWriteRequest_ProtoReflect(t *testing.T) {
	wr := &WriteRequest{}
	msg := wr.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestWriteRequest_ProtoReflect_NilReceiver(t *testing.T) {
	var wr *WriteRequest
	msg := wr.ProtoReflect()
	assert.NotNil(t, msg)
}

func TestWriteRequest_Descriptor(t *testing.T) {
	desc, path := (*WriteRequest)(nil).Descriptor()
	assert.NotNil(t, desc)
	assert.Equal(t, []int{0}, path)
}

func TestWriteRequest_MultipleTimeseries(t *testing.T) {
	wr := &WriteRequest{
		Timeseries: []*TimeSeries{
			{Labels: []*Label{{Name: "n1", Value: "v1"}}, Samples: []*Sample{{Value: 1.0, Timestamp: 1}}},
			{Labels: []*Label{{Name: "n2", Value: "v2"}}, Samples: []*Sample{{Value: 2.0, Timestamp: 2}}},
			{Labels: []*Label{{Name: "n3", Value: "v3"}}, Samples: []*Sample{{Value: 3.0, Timestamp: 3}}},
		},
	}
	assert.Equal(t, 3, len(wr.GetTimeseries()))
}

// ======================== Proto Marshal/Unmarshal Tests ========================

func TestWriteRequest_MarshalUnmarshal(t *testing.T) {
	original := &WriteRequest{
		Timeseries: []*TimeSeries{
			{
				Labels: []*Label{
					{Name: "__name__", Value: "huaweicloud_sys_ecs_cpu_util"},
					{Name: "instance_id", Value: "i-001"},
				},
				Samples: []*Sample{
					{Value: 85.5, Timestamp: 1700000000000},
				},
			},
		},
	}

	data, err := proto.Marshal(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	unmarshaled := &WriteRequest{}
	err = proto.Unmarshal(data, unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(unmarshaled.Timeseries))
	assert.Equal(t, 2, len(unmarshaled.Timeseries[0].Labels))
	assert.Equal(t, "huaweicloud_sys_ecs_cpu_util", unmarshaled.Timeseries[0].Labels[0].Value)
	assert.Equal(t, 85.5, unmarshaled.Timeseries[0].Samples[0].Value)
	assert.Equal(t, int64(1700000000000), unmarshaled.Timeseries[0].Samples[0].Timestamp)
}

func TestWriteRequest_MarshalEmpty(t *testing.T) {
	wr := &WriteRequest{}
	data, err := proto.Marshal(wr)
	assert.NoError(t, err)

	unmarshaled := &WriteRequest{}
	err = proto.Unmarshal(data, unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(unmarshaled.Timeseries))
}

func TestTimeSeries_MarshalUnmarshal(t *testing.T) {
	original := &TimeSeries{
		Labels: []*Label{
			{Name: "host", Value: "server01"},
			{Name: "region", Value: "cn-north-4"},
		},
		Samples: []*Sample{
			{Value: 42.0, Timestamp: 1000},
			{Value: 43.5, Timestamp: 2000},
		},
	}

	data, err := proto.Marshal(original)
	assert.NoError(t, err)

	unmarshaled := &TimeSeries{}
	err = proto.Unmarshal(data, unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(unmarshaled.Labels))
	assert.Equal(t, 2, len(unmarshaled.Samples))
	assert.Equal(t, "server01", unmarshaled.Labels[0].Value)
	assert.Equal(t, 43.5, unmarshaled.Samples[1].Value)
}

func TestLabel_MarshalUnmarshal(t *testing.T) {
	original := &Label{Name: "instance_id", Value: "abc-def-123"}
	data, err := proto.Marshal(original)
	assert.NoError(t, err)

	unmarshaled := &Label{}
	err = proto.Unmarshal(data, unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "instance_id", unmarshaled.Name)
	assert.Equal(t, "abc-def-123", unmarshaled.Value)
}

func TestSample_MarshalUnmarshal(t *testing.T) {
	original := &Sample{Value: 3.14159, Timestamp: 1700000000000}
	data, err := proto.Marshal(original)
	assert.NoError(t, err)

	unmarshaled := &Sample{}
	err = proto.Unmarshal(data, unmarshaled)
	assert.NoError(t, err)
	assert.InDelta(t, 3.14159, unmarshaled.Value, 0.00001)
	assert.Equal(t, int64(1700000000000), unmarshaled.Timestamp)
}

// ======================== Proto Clone/Equal Tests ========================

func TestWriteRequest_ProtoEqual(t *testing.T) {
	wr1 := &WriteRequest{
		Timeseries: []*TimeSeries{
			{Labels: []*Label{{Name: "a", Value: "b"}}, Samples: []*Sample{{Value: 1.0, Timestamp: 1}}},
		},
	}
	wr2 := &WriteRequest{
		Timeseries: []*TimeSeries{
			{Labels: []*Label{{Name: "a", Value: "b"}}, Samples: []*Sample{{Value: 1.0, Timestamp: 1}}},
		},
	}
	assert.True(t, proto.Equal(wr1, wr2))

	wr2.Timeseries[0].Samples[0].Value = 2.0
	assert.False(t, proto.Equal(wr1, wr2))
}

// ======================== FileDescriptor Test ========================

func TestFileRemoteProto_Descriptor(t *testing.T) {
	assert.NotNil(t, File_remote_proto)
	assert.Equal(t, "remote.proto", File_remote_proto.Path())
	assert.Equal(t, "prometheus", string(File_remote_proto.Package()))
}

func TestFileRemoteProto_Messages(t *testing.T) {
	msgs := File_remote_proto.Messages()
	assert.Equal(t, 4, msgs.Len())

	names := make([]string, 0, 4)
	for i := 0; i < msgs.Len(); i++ {
		names = append(names, string(msgs.Get(i).Name()))
	}
	assert.Contains(t, names, "WriteRequest")
	assert.Contains(t, names, "TimeSeries")
	assert.Contains(t, names, "Label")
	assert.Contains(t, names, "Sample")
}

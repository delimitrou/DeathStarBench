import json
import pandas as pd
import re


def collect_trace_data(path):
    with open(path) as f:
        res = json.load(f)["data"]
    
    service_id_mapping = (
        pd.json_normalize(res)
        .filter(regex="serviceName|traceID|tags")
        .rename(
            columns=lambda x: re.sub(
                r"processes\.(.*)\.serviceName|processes\.(.*)\.tags",
                lambda match_obj: match_obj.group(1)
                if match_obj.group(1)
                else f"{match_obj.group(2)}Pod",
                x,
            )
        )
        .rename(columns={"traceID": "traceId"})
    )

    service_id_mapping = (
        service_id_mapping.filter(regex=".*Pod")
        .applymap(
            lambda x: [v["value"] for v in x if v["key"] == "hostname"][0]
            if isinstance(x, list)
            else ""
        )
        .combine_first(service_id_mapping)
    )
    spans_data = pd.json_normalize(res, record_path="spans")[
        [
            "traceID",
            "spanID",
            "operationName",
            "duration",
            "processID",
            "references",
            "startTime",
        ]
    ]

    spans_with_parent = spans_data[~(spans_data["references"].astype(str) == "[]")]
    root_spans = spans_data[(spans_data["references"].astype(str) == "[]")]
    root_spans = root_spans.rename(
        columns={
            "traceID": "traceId",
            "startTime": "traceTime",
            "duration": "traceLatency"
        }
    )[["traceId", "traceTime", "traceLatency"]]
    spans_with_parent.loc[:, "parentId"] = spans_with_parent["references"].map(
        lambda x: x[0]["spanID"]
    )
    temp_parent_spans = spans_data[
        ["traceID", "spanID", "operationName", "duration", "processID"]
    ].rename(
        columns={
            "spanID": "parentId",
            "processID": "parentProcessId",
            "operationName": "parentOperation",
            "duration": "parentDuration",
            "traceID": "traceId",
        }
    )
    temp_children_spans = spans_with_parent[
        [
            "operationName",
            "duration",
            "parentId",
            "traceID",
            "spanID",
            "processID",
            "startTime",
        ]
    ].rename(
        columns={
            "spanID": "childId",
            "processID": "childProcessId",
            "operationName": "childOperation",
            "duration": "childDuration",
            "traceID": "traceId",
        }
    )
    # A merged data frame that build relationship of different spans
    merged_df = pd.merge(
        temp_parent_spans, temp_children_spans, on=["parentId", "traceId"]
    )

    merged_df = merged_df[
        [
            "traceId",
            "childOperation",
            "childDuration",
            "parentOperation",
            "parentDuration",
            "parentId",
            "childId",
            "parentProcessId",
            "childProcessId",
            "startTime",
        ]
    ]

    # Map each span's processId to its microservice name
    merged_df = merged_df.merge(service_id_mapping, on="traceId")
    merged_df = merged_df.merge(root_spans, on="traceId")
    merged_df = merged_df.assign(
        childMS=merged_df.apply(lambda x: x[x["childProcessId"]], axis=1),
        childPod=merged_df.apply(lambda x: x[f"{str(x['childProcessId'])}Pod"], axis=1),
        parentMS=merged_df.apply(lambda x: x[x["parentProcessId"]], axis=1),
        parentPod=merged_df.apply(
            lambda x: x[f"{str(x['parentProcessId'])}Pod"], axis=1
        ),
        endTime=merged_df["startTime"] + merged_df["childDuration"],
    )
    merged_df = merged_df[
        [
            "traceId",
            "traceTime",
            "startTime",
            "endTime",
            "parentId",
            "childId",
            "childOperation",
            "parentOperation",
            "childMS",
            "childPod",
            "parentMS",
            "parentPod",
            "parentDuration",
            "childDuration",
        ]
    ]
    print(merged_df)

    merged_df.to_csv('trace.csv', index=False)

if __name__ == "__main__":
    """
    traces-1695947455229.json is the metrics downloading from jaeger, this trace data will be preprocessed and saved as trace.csv
    """
    collect_trace_data('traces-1695947455229.json') 

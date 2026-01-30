#!/usr/bin/env python3
"""
Entity Extraction Spark Job
Extracts entities from archived conversations at scale
"""

import argparse
import json
import sys
from datetime import datetime

from pyspark.sql import SparkSession
from pyspark.sql.functions import (
    col, explode, struct, count, avg, max as spark_max,
    min as spark_min, collect_list, udf
)
from pyspark.sql.types import (
    StructType, StructField, StringType, IntegerType,
    FloatType, ArrayType, TimestampType
)


def create_spark_session(app_name: str) -> SparkSession:
    """Create and configure Spark session"""
    return SparkSession.builder \
        .appName(app_name) \
        .config("spark.sql.adaptive.enabled", "true") \
        .config("spark.sql.adaptive.coalescePartitions.enabled", "true") \
        .getOrCreate()


def load_conversations(spark: SparkSession, input_path: str):
    """Load conversations from data lake"""
    # Read JSON files with schema inference
    df = spark.read.json(input_path)

    return df


def extract_entities(df):
    """Extract and flatten entities from conversations"""
    # Explode entities array
    entities_df = df.select(
        col("conversation_id"),
        col("user_id"),
        col("started_at"),
        explode(col("entities")).alias("entity")
    )

    # Flatten entity structure
    flattened = entities_df.select(
        col("conversation_id"),
        col("user_id"),
        col("started_at"),
        col("entity.entity_id").alias("entity_id"),
        col("entity.entity_type").alias("entity_type"),
        col("entity.name").alias("name"),
        col("entity.value").alias("value"),
        col("entity.confidence").alias("confidence"),
        col("entity.first_seen").alias("first_seen")
    )

    return flattened


def aggregate_entity_stats(entities_df):
    """Aggregate statistics for each entity"""
    # Group by entity and calculate stats
    stats_df = entities_df.groupBy("entity_id", "entity_type", "name") \
        .agg(
            count("*").alias("mention_count"),
            collect_list("conversation_id").alias("conversations"),
            avg("confidence").alias("avg_confidence"),
            spark_max("confidence").alias("max_confidence"),
            spark_min("first_seen").alias("first_seen"),
            spark_max("first_seen").alias("last_seen")
        )

    return stats_df


def calculate_entity_importance(stats_df):
    """Calculate entity importance scores"""
    # Importance = log(mention_count) * avg_confidence
    # Simple scoring formula, can be enhanced
    from pyspark.sql.functions import log, when

    importance_df = stats_df.withColumn(
        "importance_score",
        when(col("mention_count") > 0,
             log(col("mention_count") + 1) * col("avg_confidence"))
        .otherwise(0)
    )

    return importance_df


def save_results(df, output_path: str, format: str = "parquet"):
    """Save results to data lake"""
    df.write \
        .mode("overwrite") \
        .format(format) \
        .partitionBy("entity_type") \
        .save(output_path)


def main():
    parser = argparse.ArgumentParser(description="Entity Extraction Spark Job")
    parser.add_argument("--input-path", required=True, help="Input path (conversations)")
    parser.add_argument("--output-path", required=True, help="Output path for results")
    parser.add_argument("--job-type", required=True, help="Job type")
    parser.add_argument("--start-date", help="Start date filter (YYYY-MM-DD)")
    parser.add_argument("--end-date", help="End date filter (YYYY-MM-DD)")
    parser.add_argument("--options", help="Additional options (JSON)")

    args = parser.parse_args()

    # Create Spark session
    spark = create_spark_session(f"HelixAgent-EntityExtraction-{datetime.now().isoformat()}")

    try:
        print(f"Loading conversations from {args.input_path}")
        conversations_df = load_conversations(spark, args.input_path)

        # Filter by date range if specified
        if args.start_date and args.end_date:
            conversations_df = conversations_df.filter(
                (col("started_at") >= args.start_date) &
                (col("started_at") <= args.end_date)
            )

        print("Extracting entities...")
        entities_df = extract_entities(conversations_df)

        print("Aggregating entity statistics...")
        stats_df = aggregate_entity_stats(entities_df)

        print("Calculating importance scores...")
        importance_df = calculate_entity_importance(stats_df)

        # Show sample results
        print("\nSample Results:")
        importance_df.orderBy(col("importance_score").desc()).show(10)

        print(f"\nSaving results to {args.output_path}")
        save_results(importance_df, args.output_path)

        # Print summary statistics
        total_entities = importance_df.count()
        total_mentions = importance_df.agg({"mention_count": "sum"}).collect()[0][0]

        print("\n=== Job Summary ===")
        print(f"Total unique entities: {total_entities}")
        print(f"Total entity mentions: {total_mentions}")
        print(f"Output path: {args.output_path}")

        # Write job metadata
        metadata = {
            "job_type": args.job_type,
            "input_path": args.input_path,
            "output_path": args.output_path,
            "total_entities": int(total_entities),
            "total_mentions": int(total_mentions),
            "completed_at": datetime.now().isoformat()
        }

        metadata_path = f"{args.output_path}/_metadata.json"
        spark.sparkContext.parallelize([json.dumps(metadata)]) \
            .saveAsTextFile(metadata_path)

        print("Job completed successfully!")

    except Exception as e:
        print(f"ERROR: {str(e)}", file=sys.stderr)
        raise
    finally:
        spark.stop()


if __name__ == "__main__":
    main()

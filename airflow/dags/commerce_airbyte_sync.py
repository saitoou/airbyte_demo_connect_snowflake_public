from datetime import timedelta

import pendulum
from airflow.sdk import DAG
from airflow.providers.airbyte.operators.airbyte import (
    AirbyteTriggerSyncOperator,
)
from airflow.providers.standard.operators.bash import BashOperator


AIRBYTE_CONNECTION_ID = "<snowflake-connect-id>"

with DAG(
    dag_id="commerce_airbyte_sync",
    description="Sync commerce data from PostgreSQL to Snowflake via Airbyte",

    start_date=pendulum.datetime(
        2026,
        8,
        4,
        tz="Asia/Tokyo",
    ),

    # 現在は手動実行
    schedule=None,

    # 過去期間のDAG Runを自動作成しない
    catchup=False,

    # 同じDAGを同時に複数実行しない
    max_active_runs=1,

    # 現在はTaskが1つなので、同時実行数も1
    max_active_tasks=1,

    # DAG全体の実行上限
    dagrun_timeout=timedelta(minutes=45),

    default_args={
        "owner": "data-platform",
        "retries": 1,
        "retry_delay": timedelta(minutes=10),
    },

    tags=[
        "commerce",
        "airbyte",
        "snowflake",
    ],
) as dag:

    sync_postgres_to_snowflake = AirbyteTriggerSyncOperator(
        task_id="sync_postgres_to_snowflake",
        airbyte_conn_id="airbyte_default",
        connection_id=AIRBYTE_CONNECTION_ID,

        # Airbyte Jobの完了まで待つ
        asynchronous=False,

        # Airbyte Jobの状態を10秒ごとに確認
        wait_seconds=10,

        # Airbyte Jobを最大30分待つ
        timeout=1800,

        # Airflow Task自体の制限時間
        execution_timeout=timedelta(minutes=35),
    )

    dbt_build = BashOperator(
        task_id="dbt_build",
        bash_command="""
            set -euo pipefail

            /opt/airflow/dbt_venv/bin/dbt build \
            --project-dir /opt/airflow/commerce_analytics \
            --profiles-dir /home/airflow/.dbt
        """,
        execution_timeout=timedelta(minutes=30),
    )

    sync_postgres_to_snowflake >> dbt_build


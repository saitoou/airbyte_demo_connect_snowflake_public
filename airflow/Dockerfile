FROM apache/airflow:3.3.0

# Airflow Providerなど、Airflow本体で使う依存関係
COPY airflow/requirements.txt /tmp/requirements-airflow.txt

RUN pip install --no-cache-dir \
    "apache-airflow==3.3.0" \
    -r /tmp/requirements-airflow.txt

# dbt・SQLFluff用の依存関係
COPY requirements-dbt.txt /tmp/requirements-dbt.txt

# Airflow本体と依存関係を分離するため、専用venvへインストール
RUN python -m venv /opt/airflow/dbt_venv \
    && /opt/airflow/dbt_venv/bin/pip install \
        --no-cache-dir \
        --upgrade pip \
    && /opt/airflow/dbt_venv/bin/pip install \
        --no-cache-dir \
        -r /tmp/requirements-dbt.txt
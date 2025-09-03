use criterion::{criterion_group, criterion_main, Criterion};
use renclave_enclave::seed_generator::SeedGenerator;
use std::sync::Arc;
use tokio::runtime::Runtime;

async fn benchmark_concurrent_seed_generation(concurrency: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());
    let mut handles = vec![];

    for _ in 0..concurrency {
        let generator = Arc::clone(&seed_generator);
        let handle = tokio::spawn(async move { generator.generate_seed(256, None).await.unwrap() });
        handles.push(handle);
    }

    let _results: Vec<_> = futures::future::join_all(handles).await;
}

async fn benchmark_concurrent_seed_validation(concurrency: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());
    let valid_seed = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";
    let mut handles = vec![];

    for _ in 0..concurrency {
        let generator = Arc::clone(&seed_generator);
        let seed = valid_seed.to_string();
        let handle = tokio::spawn(async move { generator.validate_seed(&seed).await.unwrap() });
        handles.push(handle);
    }

    let _results: Vec<_> = futures::future::join_all(handles).await;
}

async fn benchmark_mixed_concurrent_operations(concurrency: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());

    // Run generation and validation operations separately to avoid type conflicts
    let mut generation_handles = vec![];
    let mut validation_handles = vec![];

    for i in 0..concurrency {
        let generator = Arc::clone(&seed_generator);
        if i % 2 == 0 {
            let handle =
                tokio::spawn(async move { generator.generate_seed(256, None).await.unwrap() });
            generation_handles.push(handle);
        } else {
            let valid_seed = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";
            let handle =
                tokio::spawn(async move { generator.validate_seed(valid_seed).await.unwrap() });
            validation_handles.push(handle);
        }
    }

    // Wait for all operations to complete
    let _generation_results: Vec<_> = futures::future::join_all(generation_handles).await;
    let _validation_results: Vec<_> = futures::future::join_all(validation_handles).await;
}

fn criterion_benchmark(c: &mut Criterion) {
    let runtime = Runtime::new().unwrap();

    let mut group = c.benchmark_group("concurrent_operations");
    group.sample_size(10);
    group.measurement_time(std::time::Duration::from_secs(10));

    group.bench_function("concurrent_seed_generation_10", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_concurrent_seed_generation(10));
        });
    });

    group.bench_function("concurrent_seed_generation_50", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_concurrent_seed_generation(50));
        });
    });

    group.bench_function("concurrent_seed_validation_10", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_concurrent_seed_validation(10));
        });
    });

    group.bench_function("concurrent_seed_validation_50", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_concurrent_seed_validation(50));
        });
    });

    group.bench_function("mixed_concurrent_operations_20", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_mixed_concurrent_operations(20));
        });
    });

    group.finish();
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);

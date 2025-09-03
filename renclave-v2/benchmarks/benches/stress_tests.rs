use criterion::{criterion_group, criterion_main, Criterion};
use renclave_enclave::seed_generator::SeedGenerator;
use std::sync::Arc;
use tokio::runtime::Runtime;

async fn benchmark_stress_seed_generation_burst(operations: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());
    let mut handles = vec![];

    // Burst of operations
    for _ in 0..operations {
        let generator = Arc::clone(&seed_generator);
        let handle = tokio::spawn(async move { generator.generate_seed(256, None).await.unwrap() });
        handles.push(handle);
    }

    let _results: Vec<_> = futures::future::join_all(handles).await;
}

async fn benchmark_stress_mixed_operations_sustained(operations: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());
    let valid_seed = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

    for _ in 0..operations {
        // Generate new seed
        let _seed = seed_generator.generate_seed(256, None).await.unwrap();

        // Validate existing seed
        let _is_valid = seed_generator.validate_seed(valid_seed).await.unwrap();

        // Generate seed with passphrase
        let _seed_with_pass = seed_generator
            .generate_seed(128, Some("stress-test-passphrase"))
            .await
            .unwrap();
    }
}

async fn benchmark_stress_memory_pressure(iterations: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());
    let mut seeds = Vec::new();

    for _ in 0..iterations {
        // Generate seeds and keep them in memory
        let seed = seed_generator.generate_seed(256, None).await.unwrap();
        seeds.push(seed);

        // Validate all previously generated seeds
        for seed in &seeds {
            let _is_valid = seed_generator.validate_seed(&seed.phrase).await.unwrap();
        }

        // Limit memory usage by keeping only last 100 seeds
        if seeds.len() > 100 {
            seeds.drain(0..seeds.len() - 100);
        }
    }
}

async fn benchmark_stress_error_conditions(iterations: usize) {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());

    for _ in 0..iterations {
        // Test invalid strength
        let _result = seed_generator.generate_seed(0, None).await;

        // Test invalid seed validation
        let _result = seed_generator.validate_seed("").await;
        let _result = seed_generator.validate_seed("invalid words").await;

        // Test valid operations to maintain balance
        let _seed = seed_generator.generate_seed(256, None).await.unwrap();
        let _is_valid = seed_generator.validate_seed("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about").await.unwrap();
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    let runtime = Runtime::new().unwrap();

    let mut group = c.benchmark_group("stress_tests");
    group.sample_size(10); // Need at least 10 samples for criterion
    group.measurement_time(std::time::Duration::from_secs(15));

    group.bench_function("burst_100_operations", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_stress_seed_generation_burst(100));
        });
    });

    group.bench_function("burst_500_operations", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_stress_seed_generation_burst(500));
        });
    });

    group.bench_function("sustained_mixed_100", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_stress_mixed_operations_sustained(100));
        });
    });

    group.bench_function("memory_pressure_50", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_stress_memory_pressure(50));
        });
    });

    group.bench_function("error_conditions_100", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_stress_error_conditions(100));
        });
    });

    group.finish();
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);

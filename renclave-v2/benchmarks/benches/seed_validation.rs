use criterion::{criterion_group, criterion_main, Criterion};
use renclave_enclave::seed_generator::SeedGenerator;
use tokio::runtime::Runtime;

async fn benchmark_seed_validation_valid() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    let valid_seed = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

    for _ in 0..1000 {
        let _is_valid = seed_generator.validate_seed(valid_seed).await.unwrap();
    }
}

async fn benchmark_seed_validation_invalid() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    let invalid_seed = "invalid seed phrase that should fail validation";

    for _ in 0..1000 {
        let _is_valid = seed_generator.validate_seed(invalid_seed).await.unwrap();
    }
}

async fn benchmark_seed_validation_with_passphrase() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    let seed_with_passphrase = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about test-passphrase";

    for _ in 0..1000 {
        let _is_valid = seed_generator
            .validate_seed(seed_with_passphrase)
            .await
            .unwrap();
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    let runtime = Runtime::new().unwrap();

    let mut group = c.benchmark_group("seed_validation");
    group.sample_size(10);
    group.measurement_time(std::time::Duration::from_secs(10));

    group.bench_function("valid_seed", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_seed_validation_valid());
        });
    });

    group.bench_function("invalid_seed", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_seed_validation_invalid());
        });
    });

    group.bench_function("seed_with_passphrase", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_seed_validation_with_passphrase());
        });
    });

    group.finish();
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);

use criterion::{criterion_group, criterion_main, Criterion};
use renclave_enclave::seed_generator::SeedGenerator;
use tokio::runtime::Runtime;

async fn benchmark_seed_generation_128() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    for _ in 0..100 {
        let _seed = seed_generator.generate_seed(128, None).await.unwrap();
    }
}

async fn benchmark_seed_generation_256() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    for _ in 0..100 {
        let _seed = seed_generator.generate_seed(256, None).await.unwrap();
    }
}

async fn benchmark_seed_generation_with_passphrase() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    for _ in 0..100 {
        let _seed = seed_generator
            .generate_seed(256, Some("test-passphrase"))
            .await
            .unwrap();
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    let runtime = Runtime::new().unwrap();

    let mut group = c.benchmark_group("seed_generation");
    group.sample_size(10);
    group.measurement_time(std::time::Duration::from_secs(10));

    group.bench_function("128_bits", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_seed_generation_128());
        });
    });

    group.bench_function("256_bits", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_seed_generation_256());
        });
    });

    group.bench_function("256_bits_with_passphrase", |b| {
        b.iter(|| {
            runtime.block_on(benchmark_seed_generation_with_passphrase());
        });
    });

    group.finish();
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);

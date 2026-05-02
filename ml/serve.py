"""
SignalRoot ML Service — FastAPI server for embeddings, similarity search, and incident DNA.
"""
import hashlib
import json
import logging
import os
import time
from datetime import datetime
from typing import Optional

import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("signalroot.ml")

app = FastAPI(
    title="SignalRoot ML Service",
    description="Embedding generation, similarity search, and incident DNA computation",
    version="0.1.0",
)

# ─── Configuration ───────────────────────────────────────────────────────────
QDRANT_URL = os.getenv("QDRANT_URL", "http://localhost:6333")
QDRANT_COLLECTION = os.getenv("QDRANT_COLLECTION", "incident_dna")
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "all-MiniLM-L6-v2")
EMBEDDING_DIM = 384  # all-MiniLM-L6-v2 output dimension

# ─── Lazy-loaded model ──────────────────────────────────────────────────────
_model = None


def get_model():
    global _model
    if _model is None:
        try:
            from sentence_transformers import SentenceTransformer
            logger.info(f"Loading embedding model: {EMBEDDING_MODEL}")
            _model = SentenceTransformer(EMBEDDING_MODEL)
            logger.info("Model loaded successfully")
        except ImportError:
            logger.warning("sentence-transformers not installed, using random embeddings")
            _model = "mock"
    return _model


# ─── Pydantic models ────────────────────────────────────────────────────────
class EmbeddingRequest(BaseModel):
    text: str
    incident_id: str
    org_id: str


class EmbeddingResponse(BaseModel):
    incident_id: str
    vector_id: str
    dimensions: int


class DNARequest(BaseModel):
    incident_id: str
    org_id: str
    title: str
    summary: Optional[str] = None
    severity: str = "unknown"
    services_affected: list[str] = []
    environments: list[str] = []
    tags: list[str] = []
    detected_at: str
    resolved_at: Optional[str] = None
    signal_count: int = 0
    has_deployment_signal: bool = False
    has_dependency_signal: bool = False


class DNAResponse(BaseModel):
    incident_id: str
    feature_vector: dict


class SimilarityRequest(BaseModel):
    incident_id: str
    org_id: str
    environment: Optional[str] = None
    top_k: int = 5


class SimilarityMatch(BaseModel):
    incident_id: str
    score: float
    reason: str


class SimilarityResponse(BaseModel):
    incident_id: str
    matches: list[SimilarityMatch]
    message: Optional[str] = None


# ─── Endpoints ───────────────────────────────────────────────────────────────
@app.get("/healthz")
def health():
    return {"status": "ok"}


@app.get("/readyz")
def ready():
    return {"status": "ready", "model_loaded": _model is not None}


@app.post("/api/v1/embed", response_model=EmbeddingResponse)
def generate_embedding(req: EmbeddingRequest):
    """Generate an embedding vector for incident text and store in Qdrant."""
    start = time.time()

    # Truncate to 2000 tokens (~8000 chars as rough estimate)
    text = req.text[:8000] if len(req.text) > 8000 else req.text
    if len(req.text) > 8000:
        logger.warning(f"Text truncated for incident {req.incident_id}: {len(req.text)} chars -> 8000")

    model = get_model()
    if model == "mock":
        # Generate deterministic mock embedding from text hash
        text_hash = hashlib.sha256(text.encode()).digest()
        rng = np.random.RandomState(int.from_bytes(text_hash[:4], "big"))
        vector = rng.randn(EMBEDDING_DIM).astype(float)
    else:
        vector = model.encode(text)

    # Normalize
    norm = np.linalg.norm(vector)
    if norm > 0:
        vector = vector / norm

    vector_id = f"{req.org_id}_{req.incident_id}"

    # Store in Qdrant
    try:
        _store_vector(vector_id, vector.tolist(), {
            "incident_id": req.incident_id,
            "org_id": req.org_id,
        })
    except Exception as e:
        logger.error(f"Failed to store vector in Qdrant: {e}")
        # Non-fatal — we still return the vector_id

    duration = time.time() - start
    logger.info(f"Embedding generated for {req.incident_id} in {duration:.3f}s")

    return EmbeddingResponse(
        incident_id=req.incident_id,
        vector_id=vector_id,
        dimensions=EMBEDDING_DIM,
    )


@app.post("/api/v1/dna", response_model=DNAResponse)
def compute_dna(req: DNARequest):
    """Compute structured feature vector (Incident DNA) for explainability."""
    detected = datetime.fromisoformat(req.detected_at.replace("Z", "+00:00"))
    hour = detected.hour

    if hour < 6:
        time_bucket = 0  # night
    elif hour < 12:
        time_bucket = 1  # morning
    elif hour < 18:
        time_bucket = 2  # afternoon
    else:
        time_bucket = 3  # evening

    duration_minutes = -1.0
    if req.resolved_at:
        resolved = datetime.fromisoformat(req.resolved_at.replace("Z", "+00:00"))
        duration_minutes = (resolved - detected).total_seconds() / 60.0

    severity_map = {"critical": 1.0, "high": 0.75, "medium": 0.5, "low": 0.25, "info": 0.1}
    severity_score = severity_map.get(req.severity, 0.0)

    service_fp = hashlib.sha256(",".join(sorted(req.services_affected)).encode()).hexdigest()[:16]

    feature_vector = {
        "time_of_day_bucket": time_bucket,
        "day_of_week": detected.weekday(),
        "duration_minutes": duration_minutes,
        "severity_score": severity_score,
        "signal_count": req.signal_count,
        "services_count": len(req.services_affected),
        "service_fingerprint": service_fp,
        "has_deployment_signal": req.has_deployment_signal,
        "has_dependency_signal": req.has_dependency_signal,
        "environment": req.environments[0] if req.environments else "unknown",
    }

    return DNAResponse(incident_id=req.incident_id, feature_vector=feature_vector)


@app.post("/api/v1/similar", response_model=SimilarityResponse)
def find_similar(req: SimilarityRequest):
    """Find similar incidents using vector similarity search."""
    try:
        matches = _search_similar(req.incident_id, req.org_id, req.environment, req.top_k)
        if not matches:
            return SimilarityResponse(
                incident_id=req.incident_id,
                matches=[],
                message="Not enough history yet" if True else None,
            )
        return SimilarityResponse(incident_id=req.incident_id, matches=matches)
    except Exception as e:
        logger.error(f"Similarity search failed: {e}")
        return SimilarityResponse(
            incident_id=req.incident_id,
            matches=[],
            message=f"Search unavailable: {str(e)}",
        )


# ─── Qdrant helpers ─────────────────────────────────────────────────────────
def _ensure_collection():
    """Create Qdrant collection if it doesn't exist."""
    import httpx
    try:
        resp = httpx.get(f"{QDRANT_URL}/collections/{QDRANT_COLLECTION}")
        if resp.status_code == 404:
            httpx.put(
                f"{QDRANT_URL}/collections/{QDRANT_COLLECTION}",
                json={
                    "vectors": {"size": EMBEDDING_DIM, "distance": "Cosine"},
                },
            )
            logger.info(f"Created Qdrant collection: {QDRANT_COLLECTION}")
    except Exception as e:
        logger.warning(f"Could not connect to Qdrant: {e}")


def _store_vector(vector_id: str, vector: list[float], payload: dict):
    """Store a vector in Qdrant."""
    import httpx
    _ensure_collection()
    httpx.put(
        f"{QDRANT_URL}/collections/{QDRANT_COLLECTION}/points",
        json={
            "points": [
                {
                    "id": hashlib.md5(vector_id.encode()).hexdigest()[:16],
                    "vector": vector,
                    "payload": payload,
                }
            ]
        },
        timeout=10.0,
    )


def _search_similar(incident_id: str, org_id: str, environment: str | None, top_k: int) -> list[SimilarityMatch]:
    """Search for similar vectors in Qdrant."""
    import httpx

    vector_id = f"{org_id}_{incident_id}"
    point_id = hashlib.md5(vector_id.encode()).hexdigest()[:16]

    # Get the query vector
    resp = httpx.get(f"{QDRANT_URL}/collections/{QDRANT_COLLECTION}/points/{point_id}")
    if resp.status_code != 200:
        return []

    query_vector = resp.json().get("result", {}).get("vector", [])
    if not query_vector:
        return []

    # Search
    filters = {"must": [{"key": "org_id", "match": {"value": org_id}}]}
    if environment:
        filters["must"].append({"key": "environment", "match": {"value": environment}})

    search_resp = httpx.post(
        f"{QDRANT_URL}/collections/{QDRANT_COLLECTION}/points/search",
        json={
            "vector": query_vector,
            "filter": filters,
            "limit": top_k + 1,  # +1 to exclude self
            "with_payload": True,
        },
        timeout=10.0,
    )

    if search_resp.status_code != 200:
        return []

    results = search_resp.json().get("result", [])
    matches = []
    for r in results:
        match_id = r.get("payload", {}).get("incident_id", "")
        if match_id == incident_id:
            continue
        matches.append(SimilarityMatch(
            incident_id=match_id,
            score=round(r.get("score", 0.0), 4),
            reason="similar signal pattern",
        ))

    return matches[:top_k]


# ─── Startup ────────────────────────────────────────────────────────────────
@app.on_event("startup")
async def startup():
    logger.info("ML service starting")
    try:
        _ensure_collection()
    except Exception:
        logger.warning("Qdrant not available at startup — will retry on first request")

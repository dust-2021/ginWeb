from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import ed25519
import base64

# 生成ED25519密钥对
private_key = ed25519.Ed25519PrivateKey.generate()
public_key = private_key.public_key()

# 获取原始32字节密钥
private_bytes = private_key.private_bytes(
    encoding=serialization.Encoding.Raw,
    format=serialization.PrivateFormat.Raw,
    encryption_algorithm=serialization.NoEncryption()
)

public_bytes = public_key.public_bytes(
    encoding=serialization.Encoding.Raw,
    format=serialization.PublicFormat.Raw
)

print(base64.b64encode(public_bytes).decode('ascii'), list(public_bytes))
print(base64.b64encode(private_bytes).decode('ascii'), list(private_bytes))

import nacl.bindings
import nacl.utils

def generate_curve25519_keypair():
    """
    生成 Curve25519 密钥对
    
    Returns:
        tuple: (private_key, public_key) 各32字节
    """
    # 生成私钥（32字节随机数）
    private_key = nacl.utils.random(nacl.bindings.crypto_box_SECRETKEYBYTES)
    
    # 从私钥计算公钥
    public_key = nacl.bindings.crypto_scalarmult_base(private_key)
    
    return private_key, public_key

# 使用示例
private_key, public_key = generate_curve25519_keypair()

print(f"私钥 ({len(private_key)} 字节): {private_key.hex()}")
print(f"公钥 ({len(public_key)} 字节): {public_key.hex()}")

# 验证长度
assert len(private_key) == 32, "私钥必须是32字节"
assert len(public_key) == 32, "公钥必须是32字节"


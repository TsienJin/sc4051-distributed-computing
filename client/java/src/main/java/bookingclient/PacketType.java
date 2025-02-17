package bookingclient;

public enum PacketType {
    ERROR((byte) 0x01),
    REQUEST((byte) 0x02),
    RESPONSE((byte) 0x03),
    ACK((byte) 0x04),
    REQUEST_RESEND((byte) 0x05);

    private final byte code;

    PacketType(byte code) {
        this.code = code;
    }

    public byte getCode() {
        return code;
    }

    public static PacketType fromCode(byte code) {
        for (PacketType type : PacketType.values()) {
            if (type.code == code) {
                return type;
            }
        }
        throw new IllegalArgumentException("Unknown packet type code: " + code);
    }
}

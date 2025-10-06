#include "SCDInfoRead.h"
#include <QFile>
#include <QByteArray>
#include <QtEndian>
#include <QDateTime>
#include <QDebug>
#include <QFileInfo>

// 定义搜狗拼音词库的偏移区间
/** 偏移区间说明：
魔法字节：0x004-0x005（44 43为搜狗输入法ng版官方词库）
词库ID：0x001C-0x005B
词库生成时间戳：0x011C-0x011F
词库词条数量：0x124-0x127
词库名称：0x130-0x337
词库类别：0x338-0x53F
词库备注：0x540-0xD3F
词库示例词：0xD40-0x153
**/

const int magicBytesStart = 0x004;
const int idStart = 0x01C;
const int idEnd = 0x05B;
const int timestampStart = 0x11C;
const int timestampEnd = 0x11F;
const int phraseCountStart = 0x124;
const int phraseCountEnd = 0x127;
const int nameStart = 0x130;
const int nameEnd = 0x337;
const int categoryStart = 0x338;
const int categoryEnd = 0x53F;
const int remarkStart = 0x540;
const int remarkEnd = 0xD3F;
const int exampleStart = 0xD40;
const int exampleEnd = 0x153F;

// 内部函数：读取 UTF-16LE 字符串
static QString readString(const QString &filePath, int offsetStart, int offsetEnd)
{
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) return {};

    int size = offsetEnd - offsetStart;
    if (size <= 0) return {};

    if (!file.seek(offsetStart)) return {};

    QByteArray byteArray = file.read(size);

    // 转为 UTF-16LE 字符串
    QString result = QString::fromUtf16(reinterpret_cast<const char16_t*>(byteArray.constData()),
                                        byteArray.size() / 2);

    result.remove(QChar(0x00));
    result = result.trimmed();
    return result;
}

// 内部函数：读取词条数
static int readPhraseCount(const QString &filePath, int offsetStart, int offsetEnd)
{
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) return 0;
    file.seek(offsetStart);
    QByteArray byteArray = file.read(offsetEnd - offsetStart);
    return static_cast<int>(qFromLittleEndian<quint32>((const uchar*)byteArray.data()));
}

// 内部函数：读取时间戳
static unsigned readTimestamp(const QString &filePath, int offsetStart, int offsetEnd)
{
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) return 0;
    file.seek(offsetStart);
    QByteArray byteArray = file.read(offsetEnd - offsetStart + 1);
    return qFromLittleEndian<quint32>((const uchar*)byteArray.data());
}

// 文件头检查实现
std::pair<bool, bool> SCDInfoRead::checkSogouHeader(const QString &filePath)
{
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        qDebug() << "无法打开文件:" << filePath;
        return {false, false};
    }

    QByteArray header = file.read(12);
    file.close();

    if (header.size() < 12) {
        qDebug() << "文件太短，无法读取完整文件头";
        return {false, false};
    }

    QByteArray prefix = header.left(4);
    QByteArray middle = header.mid(4, 2);
    QByteArray suffix = header.mid(6, 6);

    QByteArray expectedPrefix = QByteArray::fromHex("40150000");
    QByteArray expectedSuffix = QByteArray::fromHex("530101000000");

    if (prefix != expectedPrefix || suffix != expectedSuffix) {
        return {false, false};
    }

    bool isOfficial = (middle == QByteArray::fromHex("4443"));
    return {true, isOfficial};
}

// 读取词库信息
SCDInfo SCDInfoRead::readSCDInfo(const QString &filePath)
{
    SCDInfo info;

    auto [valid, isOfficial] = checkSogouHeader(filePath);
    if (!valid) {
        info.allInformation = "非法的细胞词库文件头！";
        info.isOfficial = false;
        return info;
    }
    info.isOfficial = isOfficial;

    info.id = readString(filePath, idStart, idEnd);
    info.name = readString(filePath, nameStart, nameEnd);
    info.category = readString(filePath, categoryStart, categoryEnd);
    info.remark = readString(filePath, remarkStart, remarkEnd);
    info.example = readString(filePath, exampleStart, exampleEnd);
    info.phraseCount = readPhraseCount(filePath, phraseCountStart, phraseCountEnd);
    info.timestamp = readTimestamp(filePath, timestampStart, timestampEnd);
    info.formattedTimestamp = QDateTime::fromSecsSinceEpoch(info.timestamp)
                              .toString("yyyy-MM-dd HH:mm:ss");

    QString fileName = QFileInfo(filePath).fileName(); // 获取文件名

    info.allInformation = QString(
        "词库文件: %1\n编号: %2\n名称: %3\n类别: %4\n备注: %5\n"
        "创建时间: %6\n时间戳： %7\n词条数量: %8\n示例词条: %9\n数据格式: %10")
            .arg(fileName)
            .arg(info.id)
            .arg(info.name)
            .arg(info.category)
            .arg(info.remark)
            .arg(info.formattedTimestamp)
            .arg(QString::number(info.timestamp))
            .arg(QString::number(info.phraseCount))
            .arg(info.example)
            .arg(info.isOfficial ? "官方词库" : "其他词库");

    return info;
}

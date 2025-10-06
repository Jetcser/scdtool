#ifndef SCDINFOREAD_H
#define SCDINFOREAD_H

#include <QString>
#include <utility> // std::pair

// 搜狗拼音细胞词库偏移区间
extern const int magicBytesStart;
extern const int idStart;
extern const int idEnd;
extern const int timestampStart;
extern const int timestampEnd;
extern const int phraseCountStart;
extern const int phraseCountEnd;
extern const int nameStart;
extern const int nameEnd;
extern const int categoryStart;
extern const int categoryEnd;
extern const int remarkStart;
extern const int remarkEnd;
extern const int exampleStart;
extern const int exampleEnd;

// 词库信息结构体
struct SCDInfo {
    QString id;
    QString name;
    QString category;
    QString remark;
    QString example;
    int phraseCount;
    QString formattedTimestamp;
    unsigned timestamp;
    bool isOfficial;
    QString allInformation;
};

namespace SCDInfoRead {
    // 读取词库信息
    SCDInfo readSCDInfo(const QString &scdFilePath);

    // 检查文件头合法性，返回 pair<是否合法, 是否官方>
    std::pair<bool, bool> checkSogouHeader(const QString &filePath);
}

#endif // SCDINFOREAD_H

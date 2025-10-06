#include <QApplication>
#include "SCDTool_GUI.h"

int main(int argc, char *argv[]) {
    QApplication app(argc, argv);
    SCDTool_GUI w;
    w.show();
    return app.exec();
}
